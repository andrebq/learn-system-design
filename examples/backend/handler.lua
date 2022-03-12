local handler = require("handler")
local computations = require("computations")
-- pretends that the system is doing an IO operation for 0.3 seconds
computations.slow(0.3)
handler.writeStatus(200)
handler.writeBody("from LUA!")
