package game

const (
    MinPlayersToStart  = 2
    MinStartWordLength = 2
    MaxStartWordLength = 6
    MAXIDENTIFIER      = 9999
    MINIDENTIFIER      = 1000
    MINROOMIDIDENTIFIER = 1000
    MAXROOMIDIDENTIFIER = 9999
    NORMALSTARTWORD    = "사과"
    IDFALLBACK         = "0000"

    WAITINGFORPLAYERSMSG = "플레이어를 기다리는 중..."
    AVAILABLEMSG         = "새 게임을 시작할 수 있습니다. 플레이어를 기다립니다."
    STARTMSG             = "게임이 시작되었습니다! %s님부터 시작하세요."
    ELIMINATEDMSG        = "님이 탈락했습니다. 이유 : "
    WINNERMSG            = "님이 승리했습니다!"
    EXITMSG              = "님이 게임에서 나갔습니다. 다음 차례 : "
    ALLEXITMSG           = "모든 플레이어가 나갔습니다. 새로운 플레이어를 기다립니다."
    CURRENTTURNMSG       = "님의 차례입니다."

    GAMEALREADYSTARTEDMSG = "이미 게임이 시작되었습니다."
    NOHOSTPRIVILEGESMSG  = "게임을 시작할 권한이 없습니다. 호스트만 게임을 시작할 수 있습니다."
    MINPLAYERTOSTARTMSG  = "게임을 시작하려면 최소 %d명의 플레이어가 필요합니다."
    NOTTOHANDLEPLAYMSG   = "현재 게임이 시작되지 않았으므로 단어를 제출할 수 없습니다."
    NOTCURRENTPLAYERSMSG = "현재 당신의 차례가 아닙니다."

    TYPEWORDMSG        = "단어를 입력하세요."
    MINWORDLENGTHMSG   = "단어는 최소 2자 이상이어야 합니다."
    WORDALREADYUSEDMSG = "이미 사용된 단어입니다."
    WORDNOTINDICTMSG   = "사전에 없는 단어입니다."
    WORDMISMATCHMSG    = "끝말이 맞지 않습니다."

    STARTJSONTYPE   = "start_game"
    SUBMITJSONTYPE  = "submit_word"
    RESETJSONTYPE   = "reset_game"
    WELCOMEJSONTYPE = "welcome"

    MARSHALERROR            = "marshal error"
    UNMARSHALERROR          = "unmarshal error"
    FAILSENDWELCOME         = "failed to send welcome to %s: %v"
    SUBMITPAYLOADERROR      = "invalid payload for submit_word: "
    UNKNOWNMESSAGETYPE      = "unknown message type: "
    ENDLOGMSG               = "game reset after endGame in room %d"
    RESETLOGMSG             = "game reset in room %d"
    STARTLOGMSG             = "Game started in room %d"
    BROADCASTLOGMSG         = "broadcasting state in room %d: %s"
    HOSTLOGMSG              = "Player %s is now the host(ID : %s)."
    HOSTCHANGELOGMSG        = "Host user changed to %s"
    ENTERPLAYERLOGMSG       = "Player %s Enter the Game(ID : %s)"
    EXITPLAYERLOGMSG        = "Player %s Exit the Game(ID : %s)"
    DELETEROOMLOGMSG        = "Deleting empty room %d"
    REMOVESPECTATORLOGMSG   = "Spectator %s removed(ID : %s)."
    STARTINGWORDERRORLOGMSG = "Error getting random start word."
    IDMAXATTEMPTSLOGMSG     = "Warning: generateUniqueID reached max attempts, returning fallback ID"

	IDSUFFIX                 = "#"
)