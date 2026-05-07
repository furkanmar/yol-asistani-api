-- OSRM motorcycle profile — self-contained (no lib/ requires)
-- api_version = 4 — compatible with osrm-backend 5.26+
--
-- Motor tercihlerine göre ayarlanmış:
--   * Secondary/tertiary yollara bonus (kıvrımlı, manzaralı)
--   * Motorway hafif penalize (varsa alternatif tercih edilir)
--   * Motorcycle access tag hiyerarşisi

api_version = 4

-- ──────────────────────────────────────────────
-- Inline helpers (lib/ gerektirmez)
-- ──────────────────────────────────────────────

local function Set(list)
  local s = {}
  for _, v in ipairs(list) do s[v] = true end
  return s
end

local function Sequence(list) return list end

local function find_access_tag(obj, hierarchy)
  for _, tag in ipairs(hierarchy) do
    local v = obj:get_value_by_key(tag)
    if v and v ~= '' then return v end
  end
  return nil
end

-- maxspeed tag → km/h (integer veya nil)
local function parse_maxspeed(val)
  if not val then return nil end
  local n = tonumber(val)
  if n then return n end
  local mph = val:match("^(%d+)%s*mph$")
  if mph then return math.floor(tonumber(mph) * 1.60934) end
  -- Named zones
  local named = {
    ["TR:urban"]    = 50,  ["TR:rural"]    = 90,  ["TR:motorway"] = 120,
    ["DE:urban"]    = 50,  ["DE:rural"]    = 100, ["DE:motorway"] = 130,
    ["RU:urban"]    = 60,  ["RU:rural"]    = 90,  ["RU:motorway"] = 110,
    ["FR:urban"]    = 50,  ["FR:rural"]    = 80,  ["FR:motorway"] = 130,
    ["none"]        = 140, ["walk"]        = 6,
  }
  return named[val]
end

-- ──────────────────────────────────────────────
-- Profile setup
-- ──────────────────────────────────────────────

function setup()
  return {
    properties = {
      max_speed_for_map_matching    = 180/3.6,
      weight_name                   = 'routability',
      process_call_tagless_node     = false,
      u_turn_penalty                = 20,
      continue_straight_at_waypoint = true,
      use_turn_restrictions         = true,
      left_hand_driving             = false,
      traffic_light_penalty         = 2,
    },

    default_mode     = mode.driving,
    default_speed    = 50,
    oneway_handling  = true,
    turn_penalty     = 7.5,
    motorway_penalty = 0.9,  -- motorway'i hafif cezalandır, scenic alternatif tercih edilsin

    -- Motor için hız profili
    speed_profile = {
      motorway       = 110, motorway_link  = 75,
      trunk          = 100, trunk_link     = 70,
      primary        = 90,  primary_link   = 60,
      secondary      = 75,  secondary_link = 50,  -- scenic / kıvrımlı — tercih edilir
      tertiary       = 65,  tertiary_link  = 40,  -- harika motor yolları
      unclassified   = 45,
      residential    = 30,
      living_street  = 10,
      service        = 20,
      road           = 50,
      track          = 20,
    },

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

    service_tag_forbidden = Set { 'emergency_access' },
    avoid = Set { 'area', 'toll_gantry' },

    surface_speeds = {
      -- nil = yol hızını kullan
      asphalt = nil, concrete = nil, paved = nil,
      compacted = 50, fine_gravel = 40, paving_stones = 40,
      metal = 35, bricks = 30, grass = 15, wood = 15,
      sett = 30, cobblestone = 25, unpaved = 20,
      gravel = 25, dirt = 15, ground = 20,
      mud = 5, sand = 8, ice = 5,
    },

    tracktype_speeds = {
      grade1 = 55, grade2 = 40, grade3 = 25, grade4 = 15, grade5 = 10,
    },

    smoothness_speeds = {
      intermediate = 50, bad = 25, very_bad = 10,
      horrible = 5, very_horrible = 3, impassable = 0,
    },
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
      result.barrier = true
    end
  end

  local highway = node:get_value_by_key("highway")
  if highway == "traffic_signals" then
    result.traffic_lights = true
  end
end

-- ──────────────────────────────────────────────
-- Way processing
-- ──────────────────────────────────────────────

function process_way(profile, way, result, relations)
  local highway = way:get_value_by_key('highway')
  local route   = way:get_value_by_key('route')

  if (not highway or highway == '') and (not route or route == '') then
    return
  end

  -- Access check
  local access = find_access_tag(way, profile.access_tags_hierarchy)
  if access and profile.access_tag_blacklist[access] then return end

  -- Speed lookup
  local speed = profile.speed_profile[highway]
  if not speed then return end

  result.forward_speed  = speed
  result.backward_speed = speed
  result.forward_mode   = mode.driving
  result.backward_mode  = mode.driving

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
    if s and s < result.forward_speed then
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
    if s and s < result.forward_speed then
      result.forward_speed  = s
      result.backward_speed = s
    end
  end

  -- Max speed tag
  local maxspeed = parse_maxspeed(way:get_value_by_key('maxspeed'))
  if maxspeed and maxspeed > 0 and maxspeed < result.forward_speed then
    result.forward_speed  = maxspeed
    result.backward_speed = maxspeed
  end

  -- Secondary/tertiary bonus: kıvrımlı/manzaralı yollara öncelik
  if highway == 'secondary' or highway == 'tertiary' or
     highway == 'secondary_link' or highway == 'tertiary_link' then
    result.forward_rate  = (result.forward_speed  / 3.6) * 1.05
    result.backward_rate = (result.backward_speed / 3.6) * 1.05
  end

  -- Motorway hafif penalize: scenic alternatif varsa tercih edilsin
  if highway == 'motorway' or highway == 'motorway_link' then
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
  setup        = setup,
  process_way  = process_way,
  process_node = process_node,
  process_turn = process_turn,
}
