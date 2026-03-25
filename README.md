# poggadaj
Open Source reimplementation of the Gadu-Gadu backend services written in Go

## Supported features

### Client features

| Feature                                 | Gadu-Gadu 3.x | Gadu-Gadu 6.x | Gadu-Gadu 7.0-7.1 | Gadu-Gadu 7.5-7.6 | Gadu-Gadu 7.7 |
|:----------------------------------------|:-------------:|:-------------:|:-----------------:|:-----------------:|:-------------:|
| Logging in                              |       ✅       |       ✅       |         ✅         |         ✅         |       ✅       |
| Getting statuses on log in              |       ?       |       ✅       |         ✅         |         ✅         |       ✅       |
| Adding contacts (in the same session)   |       ?       |       ✅       |         ✅         |         ✅         |       ✅       |
| Removing contacts (in the same session) |       ?       |       ✅       |         ✅         |         ✅         |       ✅       |
| Saving contacts on the server           |       ?       |       ✅       |         ✅         |         ✅         |       ✅       |
| Realtime status updates                 |       ~       |       ✅       |         ✅         |         ✅         |       ✅       |
| Simple statuses                         |       ✅       |       ✅       |         ✅         |         ✅         |       ✅       |
| Statuses with descriptions              |      N/A      |       ✅       |         ✅         |         ✅         |       ✅       |
| Status masks                            |      N/A      |       ❌       |         ❌         |         ❌         |       ❌       |
| Sending messages                        |       ❌       |       ✅       |         ✅         |         ✅         |       ✅       |
| Receiving messages                      |       ❌       |       ✅       |         ✅         |         ✅         |       ✅       |
| P2P                                     |       ❌       |       ❌       |         ❌         |         ❌         |       ❌       |
| P2P over a relay                        |       ❌       |       ❌       |         ❌         |         ❌         |       ❌       |
| Public directory                        |       ❌       |       ❌       |         ❌         |         ❌         |       ❌       |

### HTTP features

|              Feature              |              Implementation status               |
|:---------------------------------:|:------------------------------------------------:|
|       IP of the TCP server        |                        ✅                         |
|                Ads                |      ✅ (image ads missing, only HTML ones)       |
|           Registration            |          ❌ (not planned? I'm not sure)           |
| Public directory (modern clients) | ❌ (will get there once I get into newer clients) |

## Project structure
The project is consisted of a few components in different directories of this monorepo. At runtime they are orchestrated using Docker Compose.
- `poggadaj-tcp` is the main component. It manages the actual connection with Gadu-Gadu clients using its protocol, handles status updates, message sending, etc.
- `poggadaj-http` manages the HTTP APIs that the Gadu-Gadu clients use, like `appmsg`, `adserver`, etc.
- `poggadaj-web` is the temporary web frontend for this project which you can see at https://poggadaj.ovh/. Don't look at it it's terrible I will rewrite it I promise.
- `poggadaj-api` manages the database accesses for `poggadaj-web`. It's on its way to get rewritten along with `poggadaj-web`.
- `poggadaj-shared` is a Go module used for sharing code used by 2 or more components. It gets copied into every component's directory using `sync_shared.sh`.

## TODOs
- Add more client support
- Gadu-Gadu 3.0 support. It came out before reverse engineering on Gadu-Gadu was done by the libgadu team so we have to figure stuff on our own. We have the login packet implemented and that's it.
- Rewrite the website
- P2P support
- Clean up the code, most notably DBHelper, packet serialization, and remove global state
