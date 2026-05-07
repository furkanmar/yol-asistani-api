-- OSRM motorcycle profile
-- Base: car.lua defaults, tuned for motorcycles:
--   * Higher highway speeds
--   * Strong preference for secondary/tertiary (scenic, winding) roads
--   * Avoid motorways when alternatives exist (lower weight, not excluded)
--   * Lane filtering not modelled (OSRM doesn't track lane context)

api_version = 4

Set = require('lib/set')
Sequence = require('lib/sequence')
Handlers = require("lib/way_handlers")
Relations = require('lib/relations')

local find_access_tag = require("lib/access").find_access_tag
local limit = require("lib/maxspeed").limit
local CLASSES = require('lib/classes')

-- ──────────────────────────────────────────────
-- Profile setup
-- ──────────────────────────────────────────────

function setup()
  return {
    properties = {
      max_speed_for_map_matching      = 180/3.6, -- m/s
      weight_name                     = 'routability',
      process_call_tagless_node       = false,
      u_turn_penalty                  = 20,
      continue_straight_at_waypoint   = true,
      use_turn_restrictions           = true,
      left_hand_driving               = false,
      traffic_light_penalty           = 2,
    },

    default_mode      = mode.driving,
    default_speed     = 50,
    oneway_handling   = true,
    side_road_multiplier = 0.8,
    turn_penalty      = 7.5,
    speed_reduction   = 0.8,
    turn_bias         = 1.075,

    -- Speed table: motorcycle-optimised
    -- Highways faster than car; scenic roads not penalised
    speed_profile = {
      motorway        = 110,
      motorway_link   = 75,
      trunk           = 100,
      trunk_link      = 70,
      primary         = 90,
      primary_link    = 60,
      secondary       = 75,   -- scenic / winding — preferred
      secondary_link  = 50,
      tertiary        = 65,   -- great motorcycle roads
      tertiary_link   = 40,
      unclassified    = 45,
      residential     = 30,
      living_street   = 10,
      service         = 20,
      road            = 50,
      track           = 20,
    },

    -- Road classes motorcycles can use
    access_tag_whitelist = Set {
      'yes', 'motorcar', 'motor_vehicle', 'vehicle',
      'permissive', 'designated', 'motorcycle',
    },

    access_tag_blacklist = Set {
      'no', 'agricultural', 'forestry', 'emergency',
      'psv', 'customers', 'private', 'delivery',
    },

    access_tags_hierarchy = Sequence {
      'motorcycle', 'motorcar', 'motor_vehicle', 'vehicle', 'access',
    },

    service_tag_forbidden = Set {
      'emergency_access',
    },

    avoid = Set {
      'area', 'toll_gantry',
    },

    speeds = Sequence {
      find_access_tag
    },

    route_speeds = {
      ferry    = 5,
      shuttle_train = 10,
    },

    bridge_speeds = {
      movable  = 5,
    },

    surface_speeds = {
      asphalt    = nil,     -- use road default
      concrete   = nil,
      ["concrete:plates"] = nil,
      paved      = nil,
      cement     = 80,
      compacted  = 50,
      fine_gravel = 40,
      paving_stones = 40,
      metal      = 35,
      bricks     = 30,
      grass      = 15,
      wood       = 15,
      sett       = 30,
      cobblestone = 25,
      unpaved    = 20,
      gravel     = 25,
      dirt       = 15,
      ground     = 20,
      mud        = 5,
      sand       = 8,
      ice        = 5,
    },

    tracktype_speeds = {
      grade1 = 55,
      grade2 = 40,
      grade3 = 25,
      grade4 = 15,
      grade5 = 10,
    },

    smoothness_speeds = {
      intermediate   = 50,
      bad            = 25,
      very_bad       = 10,
      horrible       = 5,
      very_horrible  = 3,
      impassable     = 0,
    },

    -- Penalise motorways slightly so OSRM picks scenic routes
    -- when alternatives exist (used in rate_factor below)
    motorway_penalty = 0.9,
  }
end

-- ──────────────────────────────────────────────
-- Node processing
-- ──────────────────────────────────────────────

function process_node(profile, node, result, relations)
  local access = find_access_tag(node, profile.access_tags_hierarchy)
  if access then
    if profile.access_tag_blacklist[access] then
      result.barrier = true
    end
  else
    local barrier = node:get_value_by_key("barrier")
    if barrier then
      local bollard = node:get_value_by_key("bollard")
      local rising_bollard = bollard and bollard == "rising"
      if not profile.barrier_whitelist[barrier] and not rising_bollard then
        result.barrier = true
      end
    end
  end

  local traffic_signal = node:get_value_by_key("highway")
  if traffic_signal and traffic_signal == "traffic_signals" then
    result.traffic_lights = true
  end
end

-- ──────────────────────────────────────────────
-- Way processing
-- ──────────────────────────────────────────────

function process_way(profile, way, result, relations)
  local data = {
    highway     = way:get_value_by_key('highway'),
    bridge      = way:get_value_by_key('bridge'),
    route       = way:get_value_by_key('route'),
    leisure     = way:get_value_by_key('leisure'),
    man_made    = way:get_value_by_key('man_made'),
    railway     = way:get_value_by_key('railway'),
    amenity     = way:get_value_by_key('amenity'),
    public_transport = way:get_value_by_key('public_transport'),
    toll        = way:get_value_by_key('toll'),
  }

  if (not data.highway or data.highway == '') and
     (not data.route or data.route == '') then
    return
  end

  -- Explicit access check
  local access = find_access_tag(way, profile.access_tags_hierarchy)
  if access and profile.access_tag_blacklist[access] then
    return
  end

  -- Speed
  local speed = profile.speed_profile[data.highway]
  if not speed then
    return
  end

  result.forward_speed = speed
  result.backward_speed = speed
  result.forward_mode = mode.driving
  result.backward_mode = mode.driving

  -- Oneway
  local oneway = way:get_value_by_key('oneway')
  if oneway == 'yes' or oneway == '1' or oneway == 'true' then
    result.backward_mode = mode.inaccessible
  elseif oneway == '-1' then
    result.forward_mode = mode.inaccessible
  end

  -- Surface penalty
  local surface = way:get_value_by_key('surface')
  if surface and profile.surface_speeds[surface] then
    local s = profile.surface_speeds[surface]
    if s < result.forward_speed then
      result.forward_speed  = s
      result.backward_speed = s
    end
  end

  -- Tracktype penalty
  local tracktype = way:get_value_by_key('tracktype')
  if tracktype and profile.tracktype_speeds[tracktype] then
    local s = profile.tracktype_speeds[tracktype]
    if s < result.forward_speed then
      result.forward_speed  = s
      result.backward_speed = s
    end
  end

  -- Smoothness penalty
  local smoothness = way:get_value_by_key('smoothness')
  if smoothness and profile.smoothness_speeds[smoothness] then
    local s = profile.smoothness_speeds[smoothness]
    if smoothness == 'impassable' then
      result.forward_mode  = mode.inaccessible
      result.backward_mode = mode.inaccessible
      return
    end
    if s < result.forward_speed then
      result.forward_speed  = s
      result.backward_speed = s
    end
  end

  -- Max speed tag
  local maxspeed = limit(way:get_value_by_key('maxspeed'))
  if maxspeed and maxspeed > 0 then
    if maxspeed < result.forward_speed then
      result.forward_speed  = maxspeed
      result.backward_speed = maxspeed
    end
  end

  -- Slightly prefer scenic roads: boost secondary/tertiary weight
  if data.highway == 'secondary' or data.highway == 'tertiary' or
     data.highway == 'secondary_link' or data.highway == 'tertiary_link' then
    result.forward_rate  = (result.forward_speed  / 3.6) * 1.05
    result.backward_rate = (result.backward_speed / 3.6) * 1.05
  end

  -- Slightly penalise motorways so scenic alternatives score better
  if data.highway == 'motorway' or data.highway == 'motorway_link' then
    result.forward_rate  = (result.forward_speed  / 3.6) * profile.motorway_penalty
    result.backward_rate = (result.backward_speed / 3.6) * profile.motorway_penalty
  end

  result.name = way:get_value_by_key('name')
end

-- ──────────────────────────────────────────────
-- Turn processing
-- ──────────────────────────────────────────────

function process_turn(profile, turn)
  turn.duration = 0
  if turn.is_u_turn then
    turn.duration = turn.duration + profile.properties.u_turn_penalty
  end
end

return {
  setup           = setup,
  process_way     = process_way,
  process_node    = process_node,
  process_turn    = process_turn,
}
