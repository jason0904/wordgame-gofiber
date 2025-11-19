# 끝말잇기 게임

## 1. 프로젝트 개요
gofiber와 websocket을 사용한 간단한 끝말잇기 게임입니다. 사용자끼리 멀티로 단어를 주고받으면서 게임을 진행합니다.

## 2. Websocket 프로토콜 관련 설명

클라이언트는 아래의 엔드포인트를 통해 서버와 WebSocket 연결을 맺습니다.

-   **URL**: `ws://<서버 주소>/ws`
-   보안 연결(HTTPS)의 경우: `wss://<서버 주소>/ws`

연결이 성공적으로 이루어지면, 해당 클라이언트는 게임에 참가자로 자동 추가됩니다. 연결이 끊어지면 게임에서 자동으로 제외됩니다.

## 3. 메시지 형식

클라이언트와 서버 간의 모든 메시지는 **JSON 형식**을 따릅니다.

### 3.1. 서버 -> 클라이언트 메시지

서버는 게임의 상태가 변경될 때마다 연결된 모든 클라이언트에게 현재 게임 상태를 전송(Broadcast)합니다. 이 메시지는 항상 동일한 구조를 가집니다.

-   **구조**:

    ```json
    {
      "lastWord": "string",
      "players": ["string"],             // 표시형: "<name>#<id>" (예: "익명#8016")
      "currentTurnPlayerId": "string",   // 짧은 ID (예: "8016")
      "hostUserId": "string",            // 짧은 ID
      "isGameOver": false,
      "isStarted": true,
      "message": "string"
    }
    ```

-   **필드 설명**:
    -   `lastWord` (string): 마지막으로 입력된 유효한 단어. 게임 시작 시에는 사전DB에 있는 적당한 길이의 랜덤단어가 골라집니다.
    -   `players` (Array of strings): 현재 게임에 참가 중인 모든 플레이어의 ID 목록.
    -   `currentTurnPlayerId` (string): 현재 턴인 플레이어의 ID.
    -   `hostUserID` (string): 현재 방장인 플레이어의 ID.
    -   `isGameOver` (boolean): 게임이 종료되었는지 여부. `true`이면 더 이상 단어를 제출할 수 없습니다.
    -   `isStarted` (boolean) 게임이 시작되었는지 여부. `true`이면 그 게임에 참가할 수 없습니다.
    -   `message` (string): "다음 차례:...", "게임 종료!" 등 현재 게임 상황을 설명하는 메시지.

### 3.2. 클라이언트 -> 서버 메시지

클라이언트는 특정 행동을 하고자 할 때 서버로 메시지를 보냅니다. 모든 메시지는 `type`과 `payload` 필드를 가집니다.

-   **기본 구조**:

    ```json
    {
      "type": "string",
      "payload": "string"
    }
    ```

-   **메시지 종류**:

    1.  **`submit_word`**: 단어 제출

        플레이어가 자신의 턴에 단어를 제출할 때 사용합니다.

        -   **예시**:
            ```json
            {
              "type": "submit_word",
              "payload": "고양이"
            }
            ```
        -   **`payload`**: 제출할 단어 (string).
        -   **서버 동작**:
            -   자신의 턴이 맞는 플레이어가 보낸 요청인지 확인합니다.
            -   단어의 유효성(끝말잇기 규칙, 중복 여부)을 검사합니다.
            -   결과에 따라 게임 상태를 업데이트하고 모든 클라이언트에게 새로운 상태를 전송합니다.

    2.  **`reset_game`**: 새 게임 시작 요청

        플레이어가 현재 게임을 중단하고 새 게임을 시작하고 싶을 때 사용합니다.

        -   **예시**:
            ```json
            {
              "type": "reset_game",
              "payload": ""
            }
            ```
        -   **`payload`**: 빈 문자열(`""`) 또는 무시됩니다.
        -   **서버 동작**:
            -   현재 게임의 모든 상태(단어, 턴, 점수 등)를 초기화합니다.
            -   초기화된 새로운 게임 상태를 모든 클라이언트에게 전송합니다.

## 4. 통신 흐름 예시

### 시나리오 1: 플레이어 2명 참가 및 게임 시작

1.  **플레이어 A**가 `ws://.../ws`에 연결합니다.
    -   **서버**: 플레이어 A를 게임에 추가하고 모든 클라이언트(현재 A뿐)에게 게임 상태를 전송합니다.
    -   **클라이언트 A 수신**:
        ```json
        {
          "lastWord": "",
          "players": ["A#A의ID"],
          "currentTurnPlayerId": "<A의 ID>",
          "hostUserId": "<A의 ID>",
          "isGameOver": false,
          "isStarted": false,
          "message": "새 게임을 시작할 수 있습니다. 플레이어를 기다립니다."
        }
        ```
2.  **플레이어 B**가 `ws://.../ws`에 연결합니다.
    -   **서버**: 플레이어 B를 게임에 추가하고 모든 클라이언트(A, B)에게 게임 상태를 전송합니다.
    -   **클라이언트 A, B 수신**:
        ```json
        {
          "lastWord": "",
          "players": ["A#A의ID", "B#B의ID"],
          "currentTurnPlayerId": "<A의 ID>",
          "hostUserId": "<A의 ID>",
          "isGameOver": false,
          "isStarted": false,
          "message": "새 게임을 시작할 수 있습니다. 플레이어를 기다립니다."
        }
        ```
3.  **플레이어 A**가 게임을 시작합니다.
    -   **클라이언트 A(방장) 송신**  `{"type": "start_game"}`
    -   **서버**: 플레이어 A가 게임을 시작하고, 시작단어가 자동으로 골라집니다. 모든 클라이언트(A, B)에게 게임 상태를 전송합니다.
    -   **클라이언트 A, B 수신**:
        ```json
        {
          "lastWord": "대게",
          "players": ["A#A의ID", "B#B의ID"],
          "currentTurnPlayerId": "<A의 ID>",
          "hostUserId": "<A의 ID>",
          "isGameOver": false,
          "isStarted": true,
          "message": "게임 시작! A#A의ID님부터 시작하세요."
        }
        ```
### 시나리오 2: 정상적인 턴 진행

1.  **플레이어 A**의 턴입니다. A가 단어 "게임"을 제출합니다.
    -   **클라이언트 A 송신**: `{"type": "submit_word", "payload": "게임"}`
    -   **서버**: 단어가 유효함을 확인하고, 다음 턴을 B로 넘깁니다. 새로운 게임 상태를 A, B에게 전송합니다.
    -   **클라이언트 A, B 수신**:
        ```json
        {
          "lastWord": "게임",
          "players": ["A#A의ID", "B#B의ID"],
          "currentTurnPlayerId": "<A의 ID>",
          "hostUserId": "<A의 ID>",
          "isGameOver": false,
          "isStarted": true,
          "message": "B#B의ID님의 차례입니다.
        }
        ```

### 시나리오 3: 게임 종료

1.  **플레이어 B**의 턴입니다. `lastWord`가 "게임"인데, B가 "자동차"를 제출합니다.
    -   **클라이언트 B 송신**: `{"type": "submit_word", "payload": "자동차"}`
    -   **서버**: 단어가 규칙에 어긋남을 확인하고, 게임 종료 상태로 변경 후 A, B에게 전송합니다. 그 후 자동으로 로비로 나가집니다.
    -   **클라이언트 A, B 수신**:
        ```json
        {
          "lastWord": "게임",
          "players": ["A#A의ID", "B#B의ID"],
          "currentTurnPlayerId": "<A의 ID>",
          "hostUserId": "<A의 ID>",
          "isGameOver": false,
          "isStarted": true,
          "message": "잘못된 단어입니다! '임' (으)로 시작해야 합니다. B#B의ID님의 패배!"
        }
        ```
    -   **클라이언트 B 송신**: `{"type": "reset_game"}`
## 5. 기능 구현 목록

### 유저
- [x] 유저는 본인의 이름를 생성한다.
- [x] 유저는 랜덤ID를 부여받는다.
- [x] 유저는 방에 들어올 수 있다.
- [x] 유저는 방을 생성할 수 있다.

### 방
- [x] 유저가 들어오고 나갈 수 있다.
- [x] 방에 유저가 없으면 방은 없어진다. 
- [x] 방을 만든 유저가 방장이 된다.
- [x] 방장이 나갈 경우 남은 아무에게 방장을 양도한다.
- [x] 방장은 게임을 시작할 수 있다.

### 게임
- [x] 시작 단어를 제공한다.
- [x] 단어가 사전db에 없고, 단어의 시작단어가 전 단어의 끝단어가 아닐 경우에 탈락한다.
- [x] 가장 마지막에 남아있는 사람이 우승자다.
