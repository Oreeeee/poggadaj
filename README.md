# poggadaj
WIP Open Source reimplementation of Gadu-Gadu backend

Currently the GG60 (Gadu-Gadu 6.1) protocol is targeted

## Supported features

### Client features

| Feature                                 |  Gadu-Gadu 6.x  |
|:----------------------------------------|:---------------:|
| Logging in                              | ~ (static seed) |
| Getting statuses on log in              |        ✅        |
| Adding contacts (in tha same session)   |        ✅        |
| Removing contacts (in the same session) |        ❌        |
| Realtime status updates                 |        ✅        |
| Simple statuses                         |        ✅        |
| Statuses with descriptions              |        ✅        |
| Sending messages                        |        ✅        |
| Receiving messages                      |        ✅        |
| P2P                                     |        ❌        |
| P2P over a relay                        |        ❌        |
| Public directory                        |        ❌        |

### HTTP features

|              Feature              |                    Implementation status                    |
|:---------------------------------:|:-----------------------------------------------------------:|
|       IP of the TCP server        | ~ (Gadu-Gadu 6.x - static response only); ❌ (other clients) |
|                Ads                |                    ~ (pretty basic atm)                     |
|           Registration            |                ❌ (not planned? I'm not sure)                |
| Public directory (modern clients) |      ❌ (will get there once I get into newer clients)       |
