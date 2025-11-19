# 끝말잇기 게임

## 1. 프로젝트 개요
gofiber와 websocket을 사용한 간단한 끝말잇기 게임입니다.   
사용자 여러명이 게임방을 파서 게임을 진행하며 최종 1명이 남을때까지 게임을 진행합니다.

## 2. 게임 화면
### 로비 화면
<img width="1184" height="601" alt="Image" src="https://github.com/user-attachments/assets/2f2f43cc-2959-429d-a134-6e419c83fe6a" />

### 게임 룸 생성 화면
<img width="1210" height="737" alt="Image" src="https://github.com/user-attachments/assets/3868562b-32f1-484d-ba29-3c17dae8a8f6" />

### 게임 룸 접속 화면
<img width="1184" height="601" alt="Image" src="https://github.com/user-attachments/assets/9f0b4241-ecff-4500-ab48-985bdd9762fd" />

### 게임 진행 중 화면
<img width="819" height="651" alt="Image" src="https://github.com/user-attachments/assets/7622b703-f8ab-46e0-ae21-75ea7aa8330e" />


## 3. 실행 방법

### 3.1 로컬 실행 방법

#### 요구사항
 - go 1.24.5 이상

### 실행
``` bash
#1. 저장소 클론
git clone https://github.com/jason0904/wordgame-gofiber.git

#2. go 프로젝트 파일 정리
go mod tidy

#3. 실행
go run .
```
## 4. 기능 구현 목록

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
