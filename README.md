# vela-radar
安全资产探测插件

```lua
local rr = vela.radar{
  name = "radar",
  finger = {timeout = 500 , udp = false , fast = false},
}

local es = vela.elastic.default("vela-naabu-%s" , "$day")
rr.pipe(function(host)
  es.send(host)
end)

-- 外部API
rr.define()

-- 
rr.start()
```
