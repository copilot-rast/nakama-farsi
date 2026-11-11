local nk = require("nakama")

local M = {}
function M.match_init(context, setupstate)
  return { presences = {} }, 10, nk.json_encode({ kind = "cluster_e2e", tag = setupstate.tag or "" })
end
function M.match_join_attempt(context, dispatcher, tick, state, presence, metadata)
  return state, true
end
function M.match_join(context, dispatcher, tick, state, presences)
  for _, p in ipairs(presences) do state.presences[p.session_id] = p end
  return state
end
function M.match_leave(context, dispatcher, tick, state, presences)
  for _, p in ipairs(presences) do state.presences[p.session_id] = nil end
  return state
end
function M.match_loop(context, dispatcher, tick, state, messages)
  for _, m in ipairs(messages) do
    if m.op_code == 99 then
      dispatcher.match_kick({ m.sender })
    else
      dispatcher.broadcast_message(m.op_code, m.data, nil, m.sender)
    end
  end
  return state
end
function M.match_terminate(context, dispatcher, tick, state, grace_seconds) return state end
function M.match_signal(context, dispatcher, tick, state, data) return state, "pong:" .. context.node .. ":" .. data end
return M
