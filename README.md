# poggadaj
WIP Open Source reimplementation of Gadu-Gadu backend

## Supported features

### Client features

| Feature                                 | Gadu-Gadu 6.x | Gadu-Gadu 7.0-7.1 | Gadu-Gadu 7.5-7.6 |
|:----------------------------------------|:-------------:|:------------------|:-----------------:|
| Logging in                              |       ✅       | ✅                 |         ✅         |
| Getting statuses on log in              |       ✅       | ✅                 |         ✅         |
| Adding contacts (in the same session)   |       ✅       | ✅                 |         ✅         |
| Removing contacts (in the same session) |       ✅       | ✅                 |         ✅         |
| Saving contacts on the server           |       ❌       | ❌                 |         ❌         |
| Realtime status updates                 |       ✅       | ✅                 |         ✅         |
| Simple statuses                         |       ✅       | ✅                 |         ✅         |
| Statuses with descriptions              |       ✅       | ✅                 |         ✅         |
| Status masks                            |      N/A      | ❌                 |         ❌         |
| Sending messages                        |       ✅       | ✅                 |         ✅         |
| Receiving messages                      |       ✅       | ✅                 |         ✅         |
| P2P                                     |       ❌       | ❌                 |         ❌         |
| P2P over a relay                        |       ❌       | ❌                 |         ❌         |
| Public directory                        |       ❌       | ❌                 |         ❌         |

### HTTP features

|              Feature              |              Implementation status               |
|:---------------------------------:|:------------------------------------------------:|
|       IP of the TCP server        |                        ✅                         |
|                Ads                |      ✅ (image ads missing, only HTML ones)       |
|           Registration            |          ❌ (not planned? I'm not sure)           |
| Public directory (modern clients) | ❌ (will get there once I get into newer clients) |
