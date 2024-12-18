package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"bytes"
	"encoding/gob"
	"labrpc"
	"math/rand"
	"sync"
	"time"
)

// import "bytes"
// import "encoding/gob"

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

type LogEntry struct {
	Command interface{} // 客户端要求的指令
	Term    int         // 此日志条目的term
	Index   int         // 此日志条目的index
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int // index into peers[]

	//Persistent
	currentTerm int
	votedFor    int
	log         []LogEntry

	commitIndex int
	lastApplied int

	nextIndex  []int //记录对于集群中每个节点，下一个需要发送给该节点的日志条目的索引位置。
	matchIndex []int //记录对于集群中每个节点，目前已知该节点复制成功的日志条目中最高（最新）的索引位置。

	state          int
	electionTimer  *time.Timer
	heartbeatTimer *time.Timer
	countvote      int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isleader = rf.state == 3
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	d.Decode(&rf.currentTerm)
	d.Decode(&rf.votedFor)
	d.Decode(&rf.log)
}

// example RequestVote RPC arguments structure.
type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// example RequestVote RPC reply structure.
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm || (args.Term == rf.currentTerm && rf.votedFor != -1 && rf.votedFor != args.CandidateId) {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}
	if args.Term >= rf.currentTerm {
		rf.state = 1
		rf.currentTerm = args.Term
		rf.votedFor = args.CandidateId
		rf.electionTimer.Reset(randTime())
		reply.Term, reply.VoteGranted = rf.currentTerm, true
		return
	}
	if args.Term == rf.currentTerm {
		rf.votedFor = args.CandidateId
		rf.electionTimer.Reset(randTime())
		reply.Term, reply.VoteGranted = rf.currentTerm, true
	}

}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PreLogTerm   int
	LeaderCommit int
	Entries      []LogEntry
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}
	rf.currentTerm = args.Term
	rf.votedFor = -1
	rf.state = 1
	rf.electionTimer.Reset(randTime())
	rf.persist()

	if args.PrevLogIndex < len(rf.log) {
		reply.Term, response.Success = 0, false
		return

	}
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	return index, term, isLeader
}

// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
func (rf *Raft) Kill() {
	rf.mu.Lock()
	rf.me = -1
	rf.electionTimer.Reset(0)
	rf.mu.Unlock()
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here.
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
	rf.currentTerm = 0
	rf.state = 1
	rf.votedFor = -1
	rf.countvote = 0
	rf.log = make([]LogEntry, 1)
	rf.electionTimer = time.NewTimer(randTime())
	rf.heartbeatTimer = time.NewTimer(randTime())

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	//rf.applyCond = sync.NewCond(&rf.mu)
	//lastLog := rf.getLastLog()
	go rf.checktimeout()

	return rf
}

func (rf *Raft) checktimeout() {
	for rf.me != -1 {
		select {
		case <-rf.electionTimer.C:
			rf.mu.Lock()
			rf.state = 2
			rf.currentTerm++
			rf.Election()
			rf.electionTimer.Reset(randTime())
			rf.mu.Unlock()
		case <-rf.heartbeatTimer.C:
			rf.mu.Lock()
			if rf.state == 3 {
				//rf.BroadcastHeartbeat(true)
				rf.heartbeatTimer.Reset(randTime())
			}
			rf.mu.Unlock()
		}
	}
}

func randTime() time.Duration {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randt := time.Duration(r.Intn(150) + 350)
	return randt * time.Millisecond
}

func (rf *Raft) Election() {
	args := RequestVoteArgs{Term: rf.currentTerm, CandidateId: rf.me}
	rf.countvote = 1
	rf.votedFor = rf.me
	//rf.persist()
	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}
		go func(server int) {
			reply := RequestVoteReply{}
			if rf.state == 2 && rf.sendRequestVote(i, args, &reply) {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if rf.currentTerm == reply.Term && reply.VoteGranted {
					rf.countvote++
					if rf.countvote > len(rf.peers)/2 {
						DPrintf("{Node %v} receives majority votes in term %v", rf.me, rf.currentTerm)
						rf.state = 3
						//rf.BroadcastHeartbeat(true)
					}
				} else {
					if reply.Term > rf.currentTerm {
						rf.currentTerm = reply.Term
						rf.state = 1
						rf.votedFor = -1
						//rf.persist()
					}
				}
			}
		}(i)
	}
}
