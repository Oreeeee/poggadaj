# poggadaj
Open Source reimplementation of the Gadu-Gadu backend services written in Go

## Supported features

### Client features

| Feature                                 | Gadu-Gadu 3.x | Gadu-Gadu 4.0-4.6 | Gadu-Gadu 4.8.x | Gadu-Gadu 5.0 | Gadu-Gadu 6.x | Gadu-Gadu 7.0-7.1 | Gadu-Gadu 7.5-7.6 | Gadu-Gadu 7.7 |
| --------------------------------------- | :-----------: | -------------     | : --------:     | :-----------: | :-----------: | :---------------: | :---------------: | :-----------: |
| Logging in                              | ✅             | ✅                 | ✅               | ✅             | ✅             | ✅                 | ✅                 | ✅             |
| Getting statuses on log in              | ✅             | ❌                 | ❌               | ❌             | ✅             | ✅                 | ✅                 | ✅             |
| Adding contacts (in the same session)   | ✅             | ✅                 | ✅               | ~             | ✅             | ✅                 | ✅                 | ✅             |
| Removing contacts (in the same session) | N/A           | N/A               | N/A             | ✅             | ✅             | ✅                 | ✅                 | ✅             |
| Saving contacts on the server           | N/A           | N/A               | ❌            | ❌             | ✅             | ✅                 | ✅                 | ✅             |
| Realtime status updates                 | ✅             | ✅                 | ✅               | ❌             | ✅             | ✅                 | ✅                 | ✅             |
| Simple statuses                         | ✅             | ✅                 | ✅               | ~             | ✅             | ✅                 | ✅                 | ✅             |
| Statuses with descriptions              | N/A           | N/A               | N/A             | ?             | ✅             | ✅                 | ✅                 | ✅             |
| Status masks                            | N/A           | N/A               | N/A             | ?             | ❌             | ❌                 | ❌                 | ❌             |
| Sending messages                        | ✅             | ✅                 | ✅               | ✅             | ✅             | ✅                 | ✅                 | ✅             |
| Receiving messages                      | ✅             | ✅                 | ✅               | ✅             | ✅             | ✅                 | ✅                 | ✅             |
| P2P                                     | N/A           | N/A               | N/A             | ?             | ❌             | ❌                 | ❌                 | ❌             |
| P2P over a relay                        | N/A           | N/A               | N/A             | ?             | ❌             | ❌                 | ❌                 | ❌             |
| Public directory                        | ❌             | ❌                 | ❌               | ✅[^1]         | ✅[^1]         | ✅[^1]             | ✅[^1]             | ✅[^1]         |

[^1]: Statuses are not displayed correctly. Everything else works.

### HTTP features

|              Feature              |              Implementation status               |
|:---------------------------------:|:------------------------------------------------:|
|       IP of the TCP server        |                        ✅                         |
|                Ads                |      ✅ (image ads missing, only HTML ones)       |
|           Registration            |          ❌ (not planned? I'm not sure)           |
| Public directory (modern clients) | ❌ (will get there once I get into newer clients) |

## Project structure
The project is consisted of a few components in different directories of this monorepo. At runtime they are orchestrated using Docker Compose.
- `./src/internal` contains shared code between the services.
- `./src/cmd/poggadaj-tcp` is the main component. It manages the actual connection with Gadu-Gadu clients using its protocol, handles status updates, message sending, etc.
- `./src/cmd/poggadaj-http` manages the HTTP APIs that the Gadu-Gadu clients use, like `appmsg`, `adserver`, etc.
- `./src/cmd/poggadaj-api` manages the database accesses for `poggadaj-web`. It's on its way to get rewritten along with `poggadaj-web`.
- `poggadaj-web` is the temporary web frontend for this project which you can see at https://poggadaj.ovh/. A rewrite of it is in progress.

## TODOs
- Add more client support
- Rewrite the website
- P2P support
