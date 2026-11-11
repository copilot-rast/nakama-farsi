local nk = require("nakama")

nk.register_rpc(function(context, payload)
  local p = payload ~= "" and nk.json_decode(payload) or {}
  local id = nk.match_create("cluster_e2e", { tag = p.tag or "" })
  return nk.json_encode({ match_id = id, node = context.node })
end, "cle2e_create_match")

nk.register_rpc(function(context, payload)
  local p = nk.json_decode(payload)
  nk.notification_send(p.user_id, "cluster", { from_node = context.node }, 4100, nil, false)
  return nk.json_encode({ node = context.node })
end, "cle2e_notify")

nk.register_rpc(function(context, payload)
  local p = nk.json_decode(payload)
  local r = nk.match_signal(p.match_id, "ping")
  return nk.json_encode({ result = r, node = context.node })
end, "cle2e_signal")

nk.register_rpc(function(context, payload)
  local p = nk.json_decode(payload)
  local list = nk.match_list(100, true, nil, nil, nil, "+label.tag:" .. p.tag)
  local ids = {}
  for _, m in ipairs(list) do table.insert(ids, m.match_id) end
  return nk.json_encode({ ids = ids, node = context.node })
end, "cle2e_list")

nk.register_rpc(function(context, payload)
  local p = nk.json_decode(payload)
  local stream = { mode = 123, label = p.label }
  if p.op == "join" then
    nk.stream_user_join(p.user_id, p.session_id, stream, false, false, "")
  elseif p.op == "send" then
    nk.stream_send(stream, p.data)
  elseif p.op == "kick" then
    nk.stream_user_kick({ user_id = p.user_id, session_id = p.session_id, node = p.node }, stream)
  end
  local users = nk.stream_user_list(stream, true, true)
  return nk.json_encode({ count = nk.stream_count(stream), users = #users, node = context.node })
end, "cle2e_stream")

nk.register_rpc(function(context, payload)
  local p = nk.json_decode(payload)
  if p.op == "ban" then
    nk.users_ban_id({ p.user_id })
  elseif p.op == "unban" then
    nk.users_unban_id({ p.user_id })
  elseif p.op == "logout" then
    nk.session_logout(p.user_id)
  elseif p.op == "disconnect" then
    nk.session_disconnect(p.session_id)
  end
  return nk.json_encode({ node = context.node })
end, "cle2e_session")

nk.register_rpc(function(context, payload)
  return nk.json_encode({ node = context.node })
end, "cle2e_whoami")
