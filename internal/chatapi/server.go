package chatapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Terry-Mao/goim/internal/ai"
	iconf "github.com/Terry-Mao/goim/internal/ai/conf"
	"github.com/Terry-Mao/goim/internal/chatapi/conf"
	"github.com/Terry-Mao/goim/internal/chatapi/dao"
	"github.com/Terry-Mao/goim/internal/chatapi/handler"
	"github.com/Terry-Mao/goim/internal/chatapi/middleware"
	"github.com/Terry-Mao/goim/internal/chatapi/model"
	"github.com/Terry-Mao/goim/internal/chatapi/service"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
)

// Server is the ChatAPI server
type Server struct {
	router          *gin.Engine
	httpServer      *http.Server
	mysql           *dao.MySQL
	conf            *conf.Config
	jwt             *middleware.JWTManager
	userDAO         *dao.UserDAO
	friendDAO       *dao.FriendDAO
	messageDAO      *dao.MessageDAO
	conversationDAO *dao.ConversationDAO
	groupDAO        *dao.GroupDAO
	aiDAO           *dao.AIDAO
	pushClient      *service.PushClient
	aiBotManager    *ai.BotManager
	aiContextMgr    *ai.ContextManager
}

// NewServer creates a new ChatAPI server
func NewServer(cfg *conf.Config) (*Server, error) {
	// Initialize MySQL
	mysql, err := dao.NewMySQL(cfg.MySQL)
	if err != nil {
		return nil, fmt.Errorf("init mysql error: %w", err)
	}

	// Initialize JWT manager
	jwtMgr := middleware.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireTime)

	// Initialize DAOs
	userDAO := dao.NewUserDAO(mysql)
	friendDAO := dao.NewFriendDAO(mysql)
	messageDAO := dao.NewMessageDAO(mysql)
	conversationDAO := dao.NewConversationDAO(mysql)
	groupDAO := dao.NewGroupDAO(mysql)
	aiDAO := dao.NewAIDAO(mysql)

	// Initialize push client
	pushClient := service.NewPushClient(cfg.Logic.Endpoint)

	// Initialize AI services
	aiConfig := iconf.FromChatAPIConfig(
		cfg.AI.Provider,
		cfg.AI.APIKey,
		cfg.AI.BaseURL,
		cfg.AI.Model,
		cfg.AI.Temperature,
		cfg.AI.MaxTokens,
	)
	aiBotManager := ai.NewBotManager(aiConfig)
	aiContextMgr := ai.NewContextManager(30 * time.Minute)

	// Register default AI bots
	defaultPersonalities := ai.DefaultPersonalities()
	for botID, personality := range defaultPersonalities {
		botIDInt := int64(9000)
		switch botID {
		case "assistant":
			botIDInt = 9001
		case "companion":
			botIDInt = 9002
		case "tutor":
			botIDInt = 9003
		case "creative":
			botIDInt = 9004
		}
		bot := &ai.Bot{
			ID:          botIDInt,
			UserID:      0, // System bot
			Name:        personality.Name,
			Personality: personality,
			Model:       cfg.AI.Model,
			Temperature: cfg.AI.Temperature,
		}
		aiBotManager.RegisterBot(bot)
	}

	// Set gin mode
	if cfg.HTTP != nil {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	s := &Server{
		router:          router,
		mysql:           mysql,
		conf:            cfg,
		jwt:             jwtMgr,
		userDAO:         userDAO,
		friendDAO:       friendDAO,
		messageDAO:      messageDAO,
		conversationDAO: conversationDAO,
		groupDAO:        groupDAO,
		aiDAO:           aiDAO,
		pushClient:      pushClient,
		aiBotManager:    aiBotManager,
		aiContextMgr:    aiContextMgr,
	}

	s.setupRoutes()

	return s, nil
}

// setupRoutes sets up all routes
func (s *Server) setupRoutes() {
	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(s.conf.JWT.Secret)

	// API routes
	api := s.router.Group("/api")
	{
		// User routes (no auth required for register/login)
		user := api.Group("/user")
		{
			user.POST("/register", handler.Wrap(s.handleRegister))
			user.POST("/login", handler.Wrap(s.handleLogin))
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware.Validate())
		{
			// User routes
			user := protected.Group("/user")
			{
				user.GET("/profile", handler.Wrap(s.handleGetProfile))
				user.PUT("/profile", handler.Wrap(s.handleUpdateProfile))
				user.GET("/search", handler.Wrap(s.handleSearchUsers))
			}

			// Friend routes
			friend := protected.Group("/friend")
			{
				friend.POST("/request", handler.Wrap(s.handleFriendRequest))
				friend.POST("/accept/:id", handler.Wrap(s.handleAcceptFriend))
				friend.POST("/reject/:id", handler.Wrap(s.handleRejectFriend))
				friend.PUT("/remark", handler.Wrap(s.handleUpdateFriendRemark))
				friend.DELETE("/delete/:id", handler.Wrap(s.handleDeleteFriend))
				friend.GET("/list", handler.Wrap(s.handleGetFriends))
				friend.GET("/requests", handler.Wrap(s.handleGetFriendRequests))
			}

			// Group routes
			group := protected.Group("/group")
			{
				group.GET("/list", handler.Wrap(s.handleGetGroups))
				group.POST("/create", handler.Wrap(s.handleCreateGroup))
				group.POST("/join/:id", handler.Wrap(s.handleJoinGroup))
				group.DELETE("/leave/:id", handler.Wrap(s.handleLeaveGroup))
				group.GET("/info/:id", handler.Wrap(s.handleGetGroupInfo))
				group.GET("/members/:id", handler.Wrap(s.handleGetGroupMembers))
				group.PUT("/info/:id", handler.Wrap(s.handleUpdateGroup))
				// New: invite, kick, set admin, etc.
				group.POST("/invite/:id", handler.Wrap(s.handleInviteMember))
				group.DELETE("/kick/:id/:userId", handler.Wrap(s.handleKickMember))
				group.PUT("/role/:id/:userId", handler.Wrap(s.handleSetMemberRole))
				group.PUT("/nickname/:id/:userId", handler.Wrap(s.handleSetMemberNickname))
				group.PUT("/mute/:id/:userId", handler.Wrap(s.handleMuteMember))
				group.POST("/transfer/:id/:userId", handler.Wrap(s.handleTransferOwnership))
				group.DELETE("/:id", handler.Wrap(s.handleDeleteGroup))
			}

			// Message routes
			message := protected.Group("/message")
			{
				message.POST("/send", handler.Wrap(s.handleSendMessage))
				message.GET("/history", handler.Wrap(s.handleGetHistory))
				message.POST("/read", handler.Wrap(s.handleMarkRead))
			}

			// Conversation routes
			conversation := protected.Group("/conversation")
			{
				conversation.GET("/list", handler.Wrap(s.handleGetConversations))
			}

			// AI routes
			ai := protected.Group("/ai")
			{
				ai.GET("/bots", handler.Wrap(s.handleGetAIBots))
				ai.POST("/bot/create", handler.Wrap(s.handleCreateAIBot))
				ai.GET("/bot/:id", handler.Wrap(s.handleGetAIBot))
				ai.PUT("/bot/:id", handler.Wrap(s.handleUpdateAIBot))
				ai.DELETE("/bot/:id", handler.Wrap(s.handleDeleteAIBot))
				ai.GET("/chat", handler.Wrap(s.handleGetAIConversations))
				ai.POST("/chat/send", handler.Wrap(s.handleSendAIMessage))
			}
		}
	}

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// Start starts the server
func (s *Server) Start() error {
	addr := s.conf.HTTP.Addr
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.conf.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.conf.HTTP.WriteTimeout) * time.Second,
	}

	log.Infof("ChatAPI server starting on %s", addr)

	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ChatAPI server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("ChatAPI server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Errorf("ChatAPI server shutdown error: %v", err)
	}

	// Close MySQL
	if err := s.mysql.Close(); err != nil {
		log.Errorf("MySQL close error: %v", err)
	}

	log.Info("ChatAPI server stopped")
	return nil
}

// Handler stubs (to be implemented)

// User handlers

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname" binding:"required,max=100"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) handleRegister(c *gin.Context) (interface{}, error) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	ctx := c.Request.Context()

	// Check if username already exists
	existing, _ := s.userDAO.FindByUsername(ctx, req.Username)
	if existing != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hash, err := dao.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Nickname:     req.Nickname,
		Status:       1, // online
	}

	if err := s.userDAO.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in database
	tokenDAO := dao.NewTokenDAO(s.mysql)
	_ = tokenDAO.Create(ctx, &model.UserToken{
		UserID:    user.ID,
		Token:     token,
		DeviceID:  c.GetHeader("User-Agent"),
		Platform:  "web",
		ExpiresAt: time.Now().Add(time.Duration(s.conf.JWT.ExpireTime) * time.Hour),
	})

	return gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"token":    token,
	}, nil
}

func (s *Server) handleLogin(c *gin.Context) (interface{}, error) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	ctx := c.Request.Context()

	// Find user by username
	user, err := s.userDAO.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Verify password
	if !dao.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Generate JWT token
	token, err := s.jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in database
	tokenDAO := dao.NewTokenDAO(s.mysql)
	_ = tokenDAO.Create(ctx, &model.UserToken{
		UserID:    user.ID,
		Token:     token,
		DeviceID:  c.GetHeader("User-Agent"),
		Platform:  "web",
		ExpiresAt: time.Now().Add(time.Duration(s.conf.JWT.ExpireTime) * time.Hour),
	})

	return gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.AvatarURL.String,
		"status":   user.Status,
		"token":    token,
	}, nil
}

func (s *Server) handleGetProfile(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	user, err := s.userDAO.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Don't return password hash
	user.PasswordHash = ""

	return gin.H{
		"user_id":   user.ID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"avatar":    user.AvatarURL.String,
		"status":    user.Status,
		"signature": user.Signature.String,
	}, nil
}

func (s *Server) handleUpdateProfile(c *gin.Context) (interface{}, error) {
	return gin.H{"message": "update profile - TODO"}, nil
}

func (s *Server) handleSearchUsers(c *gin.Context) (interface{}, error) {
	keyword := c.Query("keyword")
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	ctx := c.Request.Context()
	users, err := s.userDAO.Search(ctx, keyword, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return gin.H{"users": users}, nil
}

// Friend handlers

// FriendRequestRequest represents a friend request
type FriendRequestRequest struct {
	ToUserID int64  `json:"to_user_id" binding:"required"`
	Message  string `json:"message"`
}

func (s *Server) handleFriendRequest(c *gin.Context) (interface{}, error) {
	var req FriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Can't add yourself
	if userID == req.ToUserID {
		return nil, fmt.Errorf("cannot add yourself as friend")
	}

	// Check if target user exists
	targetUser, err := s.userDAO.FindByID(ctx, req.ToUserID)
	if err != nil || targetUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if already friends
	existing, _ := s.friendDAO.FindFriendship(ctx, userID, req.ToUserID)
	if existing != nil && existing.Status == 1 {
		return nil, fmt.Errorf("already friends")
	}

	// Check if there's a pending request
	pending, _ := s.friendDAO.FindPendingFriendRequest(ctx, userID, req.ToUserID)
	if pending != nil {
		return nil, fmt.Errorf("friend request already sent")
	}

	// Create friend request
	friendReq := &model.FriendRequest{
		FromUserID: userID,
		ToUserID:   req.ToUserID,
	}
	if req.Message != "" {
		friendReq.Message = sql.NullString{String: req.Message, Valid: true}
	}

	if err := s.friendDAO.CreateFriendRequest(ctx, friendReq); err != nil {
		return nil, fmt.Errorf("failed to create friend request: %w", err)
	}

	return gin.H{"message": "friend request sent"}, nil
}

func (s *Server) handleAcceptFriend(c *gin.Context) (interface{}, error) {
	requestID := c.Param("id")
	var reqID int64
	if _, err := fmt.Sscanf(requestID, "%d", &reqID); err != nil {
		return nil, fmt.Errorf("invalid request id")
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Find the friend request
	friendReq, err := s.friendDAO.FindFriendRequest(ctx, reqID)
	if err != nil || friendReq == nil {
		return nil, fmt.Errorf("friend request not found")
	}

	// Check if the request is for this user
	if friendReq.ToUserID != userID {
		return nil, fmt.Errorf("not your friend request")
	}

	// Check if already pending
	if friendReq.Status != 1 {
		return nil, fmt.Errorf("friend request already processed")
	}

	// Update request status to accepted
	if err := s.friendDAO.UpdateFriendRequestStatus(ctx, reqID, 2); err != nil {
		return nil, fmt.Errorf("failed to update friend request: %w", err)
	}

	// Create bidirectional friendship
	// Add to user's friend list
	if err := s.friendDAO.CreateFriendship(ctx, &model.Friendship{
		UserID:   userID,
		FriendID: friendReq.FromUserID,
		Status:   1,
	}); err != nil {
		return nil, fmt.Errorf("failed to create friendship: %w", err)
	}

	// Add to sender's friend list
	if err := s.friendDAO.CreateFriendship(ctx, &model.Friendship{
		UserID:   friendReq.FromUserID,
		FriendID: userID,
		Status:   1,
	}); err != nil {
		return nil, fmt.Errorf("failed to create friendship: %w", err)
	}

	return gin.H{"message": "friend added"}, nil
}

func (s *Server) handleRejectFriend(c *gin.Context) (interface{}, error) {
	requestID := c.Param("id")
	var reqID int64
	if _, err := fmt.Sscanf(requestID, "%d", &reqID); err != nil {
		return nil, fmt.Errorf("invalid request id")
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Find the friend request
	friendReq, err := s.friendDAO.FindFriendRequest(ctx, reqID)
	if err != nil || friendReq == nil {
		return nil, fmt.Errorf("friend request not found")
	}

	// Check if the request is for this user
	if friendReq.ToUserID != userID {
		return nil, fmt.Errorf("not your friend request")
	}

	// Update request status to rejected
	if err := s.friendDAO.UpdateFriendRequestStatus(ctx, reqID, 3); err != nil {
		return nil, fmt.Errorf("failed to update friend request: %w", err)
	}

	return gin.H{"message": "friend request rejected"}, nil
}

func (s *Server) handleUpdateFriendRemark(c *gin.Context) (interface{}, error) {
	var req struct {
		FriendID  int64  `json:"friend_id" binding:"required"`
		Remark    string `json:"remark"`
		GroupName string `json:"group_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if friendship exists
	existing, err := s.friendDAO.FindFriendship(ctx, userID, req.FriendID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("friendship not found")
	}

	// Update friendship
	existing.Remark = sql.NullString{String: req.Remark, Valid: req.Remark != ""}
	existing.GroupName = req.GroupName

	if err := s.friendDAO.UpdateFriendship(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update friendship: %w", err)
	}

	return gin.H{"message": "friend remark updated"}, nil
}

func (s *Server) handleDeleteFriend(c *gin.Context) (interface{}, error) {
	friendID := c.Param("id")
	var fID int64
	if _, err := fmt.Sscanf(friendID, "%d", &fID); err != nil {
		return nil, fmt.Errorf("invalid friend id")
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Delete friendship (both directions)
	if err := s.friendDAO.DeleteFriendship(ctx, userID, fID); err != nil {
		return nil, fmt.Errorf("failed to delete friendship: %w", err)
	}
	if err := s.friendDAO.DeleteFriendship(ctx, fID, userID); err != nil {
		return nil, fmt.Errorf("failed to delete friendship: %w", err)
	}

	return gin.H{"message": "friend deleted"}, nil
}

func (s *Server) handleGetFriends(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	friends, err := s.friendDAO.GetFriends(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}

	return gin.H{"friends": friends}, nil
}

func (s *Server) handleGetFriendRequests(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	requests, err := s.friendDAO.GetFriendRequestsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friend requests: %w", err)
	}

	return gin.H{"requests": requests}, nil
}

// Group handlers

// CreateGroupRequest represents a create group request
type CreateGroupRequest struct {
	Name       string `json:"name" binding:"required,max=100"`
	MaxMembers int    `json:"max_members"`
}

func (s *Server) handleCreateGroup(c *gin.Context) (interface{}, error) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Set default max members if not provided
	if req.MaxMembers <= 0 || req.MaxMembers > 500 {
		req.MaxMembers = 500
	}

	// Create group
	group := &model.Group{
		Name:       req.Name,
		OwnerID:    userID,
		MaxMembers: req.MaxMembers,
		JoinType:   1, // Open by default
		MuteAll:    0,
	}

	if err := s.groupDAO.CreateGroup(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// Add owner as member
	if err := s.groupDAO.AddMember(ctx, group.ID, userID, 3); err != nil {
		return nil, fmt.Errorf("failed to add owner to group: %w", err)
	}

	// Create system message
	welcomeMsg := &model.Message{
		FromUserID:       userID,
		ConversationID:   group.ID,
		ConversationType: model.ConversationTypeGroup,
		MsgType:          model.MsgTypeSystem,
		Content:          fmt.Sprintf(`{"text":"群组创建成功"}`),
		Seq:              1,
	}
	s.messageDAO.CreateMessage(ctx, welcomeMsg)

	return gin.H{
		"group_id":    group.ID,
		"group_no":    group.GroupNo,
		"name":        group.Name,
		"owner_id":    group.OwnerID,
		"max_members": req.MaxMembers,
	}, nil
}

func (s *Server) handleJoinGroup(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	var gID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if group exists
	group, err := s.groupDAO.FindByID(ctx, gID)
	if err != nil || group == nil {
		return nil, fmt.Errorf("group not found")
	}

	// Check if already a member
	isMember, _ := s.groupDAO.IsMember(ctx, gID, userID)
	if isMember {
		return nil, fmt.Errorf("already a member")
	}

	// Check join type
	if group.JoinType == 3 {
		return nil, fmt.Errorf("group is closed")
	}

	// Add member
	if err := s.groupDAO.AddMember(ctx, gID, userID, 1); err != nil {
		return nil, fmt.Errorf("failed to join group: %w", err)
	}

	return gin.H{"message": "joined group successfully"}, nil
}

func (s *Server) handleLeaveGroup(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	var gID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if group exists
	group, err := s.groupDAO.FindByID(ctx, gID)
	if err != nil || group == nil {
		return nil, fmt.Errorf("group not found")
	}

	// Owner cannot leave
	if group.OwnerID == userID {
		return nil, fmt.Errorf("owner cannot leave group")
	}

	// Remove member
	if err := s.groupDAO.RemoveMember(ctx, gID, userID); err != nil {
		return nil, fmt.Errorf("failed to leave group: %w", err)
	}

	return gin.H{"message": "left group successfully"}, nil
}

func (s *Server) handleGetGroupInfo(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	var gID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	ctx := c.Request.Context()

	group, err := s.groupDAO.FindByID(ctx, gID)
	if err != nil || group == nil {
		return nil, fmt.Errorf("group not found")
	}

	// Get members count
	members, err := s.groupDAO.GetMembers(ctx, gID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	return gin.H{
		"id":           group.ID,
		"group_no":     group.GroupNo,
		"name":         group.Name,
		"avatar":       group.AvatarURL.String,
		"owner_id":     group.OwnerID,
		"max_members":  group.MaxMembers,
		"join_type":    group.JoinType,
		"mute_all":     group.MuteAll,
		"member_count": len(members),
		"members":      members,
	}, nil
}

func (s *Server) handleGetGroupMembers(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	var gID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	ctx := c.Request.Context()

	members, err := s.groupDAO.GetMembers(ctx, gID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	return gin.H{"members": members}, nil
}

func (s *Server) handleUpdateGroup(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	var gID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Find group
	group, err := s.groupDAO.FindByID(ctx, gID)
	if err != nil || group == nil {
		return nil, fmt.Errorf("group not found")
	}

	// Check if user is owner
	if group.OwnerID != userID {
		return nil, fmt.Errorf("only owner can update group")
	}

	// Update group
	group.Name = req.Name
	if req.MaxMembers > 0 {
		group.MaxMembers = req.MaxMembers
	}

	if err := s.groupDAO.UpdateGroup(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	return gin.H{
		"group_id":    group.ID,
		"group_no":    group.GroupNo,
		"name":        group.Name,
		"max_members": group.MaxMembers,
	}, nil
}

func (s *Server) handleGetGroups(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Get all groups the user is a member of
	members, err := s.groupDAO.GetGroupsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}

	// Extract groups and enrich with member counts
	result := make([]gin.H, 0, len(members))
	for _, member := range members {
		if member.Group == nil {
			continue
		}
		group := member.Group
		groupMembers, _ := s.groupDAO.GetMembers(ctx, group.ID)
		result = append(result, gin.H{
			"id":           group.ID,
			"group_no":     group.GroupNo,
			"name":         group.Name,
			"avatar":       group.AvatarURL.String,
			"owner_id":     group.OwnerID,
			"max_members":  group.MaxMembers,
			"member_count": len(groupMembers),
			"join_type":    group.JoinType,
			"created_at":   group.CreatedAt,
		})
	}

	return gin.H{"groups": result}, nil
}

// handleInviteMember invites a user to a group (by owner/admin)
func (s *Server) handleInviteMember(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	var gID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
		Role   int8  `json:"role"` // Optional: 1=member, 2=admin
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if inviter is owner or admin
	inviterRole, err := s.groupDAO.GetMemberRole(ctx, gID, userID)
	if err != nil || (inviterRole != 3 && inviterRole != 2) {
		return nil, fmt.Errorf("only owner or admin can invite members")
	}

	// Check if target user exists
	targetUser, err := s.userDAO.FindByID(ctx, req.UserID)
	if err != nil || targetUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if already a member
	isMember, _ := s.groupDAO.IsMember(ctx, gID, req.UserID)
	if isMember {
		return nil, fmt.Errorf("user is already a member")
	}

	// Add member (default role is member)
	role := req.Role
	if role == 0 {
		role = 1 // Default to member
	}
	if inviterRole == 2 && role == 3 {
		return nil, fmt.Errorf("admin cannot make someone owner")
	}

	if err := s.groupDAO.AddMember(ctx, gID, req.UserID, role); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Create system message
	welcomeMsg := &model.Message{
		FromUserID:       userID,
		ConversationID:   gID,
		ConversationType: model.ConversationTypeGroup,
		MsgType:          model.MsgTypeSystem,
		Content:          fmt.Sprintf(`{"text":"邀请 %s 加入了群聊"}`, targetUser.Nickname),
	}
	s.messageDAO.CreateMessage(ctx, welcomeMsg)

	return gin.H{"message": "member invited successfully"}, nil
}

// handleKickMember removes a member from the group (by owner/admin)
func (s *Server) handleKickMember(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	userIDStr := c.Param("userId")
	var gID, targetUserID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}
	if _, err := fmt.Sscanf(userIDStr, "%d", &targetUserID); err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if kicker is owner or admin
	kickerRole, err := s.groupDAO.GetMemberRole(ctx, gID, userID)
	if err != nil || (kickerRole != 3 && kickerRole != 2) {
		return nil, fmt.Errorf("only owner or admin can remove members")
	}

	// Check target user's role
	targetRole, err := s.groupDAO.GetMemberRole(ctx, gID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("target user not found")
	}

	// Admin cannot kick owner or other admins (only owner can)
	if kickerRole == 2 && (targetRole == 3 || targetRole == 2) {
		return nil, fmt.Errorf("admin cannot kick owner or other admins")
	}

	// Get target user info
	targetUser, _ := s.userDAO.FindByID(ctx, targetUserID)

	// Remove member
	if err := s.groupDAO.RemoveMember(ctx, gID, targetUserID); err != nil {
		return nil, fmt.Errorf("failed to remove member: %w", err)
	}

	// Create system message
	kickMsg := &model.Message{
		FromUserID:       userID,
		ConversationID:   gID,
		ConversationType: model.ConversationTypeGroup,
		MsgType:          model.MsgTypeSystem,
		Content:          fmt.Sprintf(`{"text":"%s 被移出了群聊"}`, getNickname(targetUser)),
	}
	s.messageDAO.CreateMessage(ctx, kickMsg)

	return gin.H{"message": "member removed successfully"}, nil
}

// handleSetMemberRole sets a member's role (by owner)
func (s *Server) handleSetMemberRole(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	userIDStr := c.Param("userId")
	var gID, targetUserID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}
	if _, err := fmt.Sscanf(userIDStr, "%d", &targetUserID); err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	var req struct {
		Role int8 `json:"role" binding:"required,min=1,max=3"` // 1=member, 2=admin, 3=owner
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if operator is owner
	group, err := s.groupDAO.FindByID(ctx, gID)
	if err != nil || group == nil {
		return nil, fmt.Errorf("group not found")
	}
	if group.OwnerID != userID {
		return nil, fmt.Errorf("only owner can change member roles")
	}

	// Cannot change own role
	if targetUserID == userID {
		return nil, fmt.Errorf("cannot change your own role")
	}

	// Update role
	if err := s.groupDAO.UpdateMemberRole(ctx, gID, targetUserID, req.Role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return gin.H{"message": "role updated successfully"}, nil
}

// handleSetMemberNickname sets a member's nickname in the group
func (s *Server) handleSetMemberNickname(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	userIDStr := c.Param("userId")
	var gID, targetUserID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}
	if _, err := fmt.Sscanf(userIDStr, "%d", &targetUserID); err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	var req struct {
		Nickname string `json:"nickname" binding:"required,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Can set own nickname or owner/admin can set others'
	if targetUserID != userID {
		operatorRole, err := s.groupDAO.GetMemberRole(ctx, gID, userID)
		if err != nil || (operatorRole != 3 && operatorRole != 2) {
			return nil, fmt.Errorf("can only set your own nickname")
		}
	}

	// Update nickname
	if err := s.groupDAO.UpdateMemberNickname(ctx, gID, targetUserID, req.Nickname); err != nil {
		return nil, fmt.Errorf("failed to update nickname: %w", err)
	}

	return gin.H{"message": "nickname updated successfully"}, nil
}

// handleMuteMember mutes/unmutes a member (by owner/admin)
func (s *Server) handleMuteMember(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	userIDStr := c.Param("userId")
	var gID, targetUserID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}
	if _, err := fmt.Sscanf(userIDStr, "%d", &targetUserID); err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	var req struct {
		Duration int `json:"duration"` // Duration in minutes, 0 to unmute
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if operator is owner or admin
	operatorRole, err := s.groupDAO.GetMemberRole(ctx, gID, userID)
	if err != nil || (operatorRole != 3 && operatorRole != 2) {
		return nil, fmt.Errorf("only owner or admin can mute members")
	}

	// Check target role
	targetRole, err := s.groupDAO.GetMemberRole(ctx, gID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("target user not found")
	}

	// Admin cannot mute owner
	if operatorRole == 2 && targetRole == 3 {
		return nil, fmt.Errorf("admin cannot mute owner")
	}

	var muteUntil *time.Time
	if req.Duration > 0 {
		until := time.Now().Add(time.Duration(req.Duration) * time.Minute)
		muteUntil = &until
	}

	if err := s.groupDAO.SetMemberMute(ctx, gID, targetUserID, muteUntil); err != nil {
		return nil, fmt.Errorf("failed to set mute: %w", err)
	}

	return gin.H{"message": "mute status updated successfully"}, nil
}

// handleTransferOwnership transfers group ownership to another member
func (s *Server) handleTransferOwnership(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	userIDStr := c.Param("userId")
	var gID, newOwnerID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}
	if _, err := fmt.Sscanf(userIDStr, "%d", &newOwnerID); err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if current user is owner
	group, err := s.groupDAO.FindByID(ctx, gID)
	if err != nil || group == nil {
		return nil, fmt.Errorf("group not found")
	}
	if group.OwnerID != userID {
		return nil, fmt.Errorf("only owner can transfer ownership")
	}

	// Check if new owner is a member
	isMember, err := s.groupDAO.IsMember(ctx, gID, newOwnerID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("target user is not a member")
	}

	// Get new owner info
	newOwner, _ := s.userDAO.FindByID(ctx, newOwnerID)

	// Transfer ownership
	if err := s.groupDAO.TransferOwnership(ctx, gID, newOwnerID); err != nil {
		return nil, fmt.Errorf("failed to transfer ownership: %w", err)
	}

	// Create system message
	transferMsg := &model.Message{
		FromUserID:       userID,
		ConversationID:   gID,
		ConversationType: model.ConversationTypeGroup,
		MsgType:          model.MsgTypeSystem,
		Content:          fmt.Sprintf(`{"text":"群主转让给 %s"}`, getNickname(newOwner)),
	}
	s.messageDAO.CreateMessage(ctx, transferMsg)

	return gin.H{"message": "ownership transferred successfully"}, nil
}

// handleDeleteGroup deletes a group (by owner)
func (s *Server) handleDeleteGroup(c *gin.Context) (interface{}, error) {
	groupID := c.Param("id")
	var gID int64
	if _, err := fmt.Sscanf(groupID, "%d", &gID); err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Check if user is owner
	group, err := s.groupDAO.FindByID(ctx, gID)
	if err != nil || group == nil {
		return nil, fmt.Errorf("group not found")
	}
	if group.OwnerID != userID {
		return nil, fmt.Errorf("only owner can delete group")
	}

	// Delete group
	if err := s.groupDAO.DeleteGroup(ctx, gID); err != nil {
		return nil, fmt.Errorf("failed to delete group: %w", err)
	}

	return gin.H{"message": "group deleted successfully"}, nil
}

// Helper function to get nickname
func getNickname(user *model.User) string {
	if user == nil {
		return "用户"
	}
	if user.Nickname != "" {
		return user.Nickname
	}
	return user.Username
}

// Message handlers

// SendMessageRequest represents a send message request
type SendMessageRequest struct {
	ToUserID         int64  `json:"to_user_id"`
	ToGroupID        int64  `json:"to_group_id"`
	ConversationType int8   `json:"conversation_type"` // 1=single, 2=group
	MsgType          int8   `json:"msg_type"`          // 1=text, 2=image, etc.
	Content          string `json:"content" binding:"required"`
}

func (s *Server) handleSendMessage(c *gin.Context) (interface{}, error) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Determine conversation ID and target
	var conversationID int64
	var keys []string

	switch req.ConversationType {
	case model.ConversationTypeSingle: // Single chat
		if req.ToUserID == 0 {
			return nil, fmt.Errorf("to_user_id is required for single chat")
		}
		if req.ToUserID == userID {
			return nil, fmt.Errorf("cannot send message to yourself")
		}
		conversationID = dao.GetConversationPairID(userID, req.ToUserID)
		// Get user's key for goim push
		keys = append(keys, fmt.Sprintf("user_%d", req.ToUserID))

	case model.ConversationTypeGroup: // Group chat
		if req.ToGroupID == 0 {
			return nil, fmt.Errorf("to_group_id is required for group chat")
		}
		conversationID = req.ToGroupID
		// Group chat uses room push

	default:
		return nil, fmt.Errorf("invalid conversation type")
	}

	// Get next sequence number
	seq, err := s.messageDAO.GetNextSeq(ctx, conversationID, req.ConversationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get sequence: %w", err)
	}

	// Create message
	msg := &model.Message{
		FromUserID:       userID,
		ConversationID:   conversationID,
		ConversationType: req.ConversationType,
		MsgType:          req.MsgType,
		Content:          req.Content,
		Seq:              seq,
	}

	// Save message to database
	if err := s.messageDAO.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Reload message from database to get the actual created_at timestamp
	savedMsg, err := s.messageDAO.FindByMsgID(ctx, msg.MsgID)
	if err == nil && savedMsg != nil {
		msg = savedMsg
	} else {
		// Fallback to current time if query fails
		msg.CreatedAt = time.Now()
	}

	// Prepare message for push
	pushContent := map[string]interface{}{
		"msg_id":            msg.MsgID,
		"from_user_id":      userID,
		"conversation_id":   conversationID,
		"conversation_type": req.ConversationType,
		"msg_type":          req.MsgType,
		"content":           req.Content,
		"seq":               seq,
		"created_at":        msg.CreatedAt.Unix(),
	}
	pushJSON, _ := json.Marshal(pushContent)

	// Push message through goim
	if req.ConversationType == model.ConversationTypeSingle {
		// Single chat: push to specific user
		if err := s.pushClient.PushKeys(ctx, 1001, keys, pushJSON); err != nil {
			log.Errorf("Failed to push message: %v", err)
		}
	} else if req.ConversationType == model.ConversationTypeGroup {
		// Group chat: push to room
		room := fmt.Sprintf("group://%d", req.ToGroupID)
		if err := s.pushClient.PushRoom(ctx, 1001, "group", room, pushJSON); err != nil {
			log.Errorf("Failed to push room message: %v", err)
		}
	}

	// Update conversation
	// For sender: update or create conversation entry
	conv := &model.Conversation{
		UserID:           userID,
		TargetID:         conversationID,
		ConversationType: req.ConversationType,
		LastMsgID:        sql.NullInt64{Int64: msg.ID, Valid: true},
		LastMsgContent:   sql.NullString{String: req.Content, Valid: true},
		LastMsgTime:      sql.NullTime{Time: msg.CreatedAt, Valid: true},
	}
	s.conversationDAO.UpsertConversation(ctx, conv)

	// For receiver (single chat only): increment unread and update conversation
	if req.ConversationType == model.ConversationTypeSingle {
		recvConv := &model.Conversation{
			UserID:           req.ToUserID,
			TargetID:         conversationID,
			ConversationType: req.ConversationType,
			LastMsgID:        sql.NullInt64{Int64: msg.ID, Valid: true},
			LastMsgContent:   sql.NullString{String: req.Content, Valid: true},
			LastMsgTime:      sql.NullTime{Time: msg.CreatedAt, Valid: true},
		}
		s.conversationDAO.UpsertConversation(ctx, recvConv)
		s.conversationDAO.IncrementUnread(ctx, req.ToUserID, conversationID, req.ConversationType)
	}

	return gin.H{
		"msg_id":            msg.MsgID,
		"conversation_id":   conversationID,
		"conversation_type": req.ConversationType,
		"seq":               seq,
		"created_at":        msg.CreatedAt,
	}, nil
}

func (s *Server) handleGetHistory(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	conversationIDStr := c.Query("conversation_id")
	if conversationIDStr == "" {
		return nil, fmt.Errorf("conversation_id is required")
	}
	conversationID, _ := strconv.ParseInt(conversationIDStr, 10, 64)

	convTypeStr := c.Query("conversation_type")
	if convTypeStr == "" {
		return nil, fmt.Errorf("conversation_type is required")
	}
	convType := int8(convTypeStr[0] - '0')

	lastSeqStr := c.Query("last_seq")
	lastSeq, _ := strconv.ParseInt(lastSeqStr, 10, 64)

	limitStr := c.Query("limit")
	limit := int64(50)
	if limitStr != "" {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}
	if limit > 100 {
		limit = 100
	}

	messages, err := s.messageDAO.GetMessages(ctx, conversationID, convType, int(limit), lastSeq)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return gin.H{
		"messages": messages,
		"has_more": len(messages) == int(limit),
		"user_id":  userID,
	}, nil
}

// MarkReadRequest represents a mark read request
type MarkReadRequest struct {
	ConversationID   int64 `json:"conversation_id" binding:"required"`
	ConversationType int8  `json:"conversation_type" binding:"required"`
	MsgID            int64 `json:"msg_id" binding:"required"`
}

func (s *Server) handleMarkRead(c *gin.Context) (interface{}, error) {
	var req MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Mark message as read
	if err := s.messageDAO.MarkMessagesRead(ctx, req.MsgID, userID); err != nil {
		return nil, fmt.Errorf("failed to mark read: %w", err)
	}

	// Clear unread count
	if err := s.conversationDAO.ClearUnread(ctx, userID, req.ConversationID, req.ConversationType); err != nil {
		log.Errorf("Failed to clear unread: %v", err)
	}

	return gin.H{"message": "marked as read"}, nil
}

// Conversation handlers
func (s *Server) handleGetConversations(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	conversations, err := s.conversationDAO.GetConversations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	return gin.H{"conversations": conversations}, nil
}

// AI handlers
func (s *Server) handleGetAIBots(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	// Get predefined bots and user-created bots
	personalities := ai.DefaultPersonalities()
	bots := make([]gin.H, 0)

	// Add default bots
	for id, personality := range personalities {
		botID := int64(9000)
		switch id {
		case "assistant":
			botID = 9001
		case "companion":
			botID = 9002
		case "tutor":
			botID = 9003
		case "creative":
			botID = 9004
		}
		bots = append(bots, gin.H{
			"id":          botID,
			"name":        personality.Name,
			"personality": id,
			"role":        personality.Role,
			"tone":        personality.Tone,
			"is_default":  true,
		})
	}

	// Get user's custom bots
	userBots, err := s.aiDAO.FindBotsByUser(ctx, userID)
	if err != nil {
		log.Errorf("Failed to get user bots: %v", err)
	} else {
		for _, bot := range userBots {
			bots = append(bots, gin.H{
				"id":          bot.BotID,
				"name":        bot.Name,
				"personality": bot.Personality,
				"is_default":  false,
				"model":       bot.ModelName,
			})
		}
	}

	return gin.H{"bots": bots}, nil
}

func (s *Server) handleGetAIBot(c *gin.Context) (interface{}, error) {
	idStr := c.Param("id")
	botID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bot id: %w", err)
	}

	// Check if it's a default bot
	if botID >= 9000 && botID <= 9999 {
		personalities := ai.DefaultPersonalities()
		var personalityKey string
		var personality *ai.Personality
		switch botID {
		case 9001:
			personalityKey = "assistant"
		case 9002:
			personalityKey = "companion"
		case 9003:
			personalityKey = "tutor"
		case 9004:
			personalityKey = "creative"
		}
		personality = personalities[personalityKey]
		if personality != nil {
			return gin.H{
				"id":          botID,
				"name":        personality.Name,
				"personality": personalityKey,
				"role":        personality.Role,
				"tone":        personality.Tone,
				"traits":      personality.Traits,
				"is_default":  true,
			}, nil
		}
	}

	// Get user's custom bot
	ctx := c.Request.Context()
	bot, err := s.aiDAO.FindBotByID(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	return gin.H{
		"id":          bot.BotID,
		"name":        bot.Name,
		"personality": bot.Personality,
		"model":       bot.ModelName,
		"is_default":  false,
	}, nil
}

func (s *Server) handleCreateAIBot(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	var req struct {
		Name        string  `json:"name"`
		Personality string  `json:"personality"`
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Generate bot ID
	botID := time.Now().Unix()

	bot := &model.AIBot{
		BotID:       botID,
		UserID:      userID,
		Name:        req.Name,
		Personality: req.Personality,
		ModelName:   req.Model,
		Temperature: req.Temperature,
		MaxTokens:   1000,
	}

	if err := s.aiDAO.CreateBot(ctx, bot); err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return gin.H{
		"bot_id": bot.BotID,
		"name":   bot.Name,
	}, nil
}

func (s *Server) handleUpdateAIBot(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	idStr := c.Param("id")
	botID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bot id: %w", err)
	}

	var req struct {
		Name        string  `json:"name"`
		Personality string  `json:"personality"`
		Temperature float64 `json:"temperature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	bot, err := s.aiDAO.FindBotByID(ctx, botID)
	if err != nil || bot == nil {
		return nil, fmt.Errorf("bot not found")
	}

	if bot.UserID != userID {
		return nil, fmt.Errorf("not authorized to update this bot")
	}

	bot.Name = req.Name
	bot.Personality = req.Personality
	bot.Temperature = req.Temperature

	if err := s.aiDAO.UpdateBot(ctx, bot); err != nil {
		return nil, fmt.Errorf("failed to update bot: %w", err)
	}

	return gin.H{"message": "bot updated successfully"}, nil
}

func (s *Server) handleDeleteAIBot(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	idStr := c.Param("id")
	botID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bot id: %w", err)
	}

	bot, err := s.aiDAO.FindBotByID(ctx, botID)
	if err != nil || bot == nil {
		return nil, fmt.Errorf("bot not found")
	}

	if bot.UserID != userID {
		return nil, fmt.Errorf("not authorized to delete this bot")
	}

	if err := s.aiDAO.DeleteBot(ctx, bot.ID); err != nil {
		return nil, fmt.Errorf("failed to delete bot: %w", err)
	}

	return gin.H{"message": "bot deleted successfully"}, nil
}

func (s *Server) handleGetAIConversations(c *gin.Context) (interface{}, error) {
	userID := middleware.GetUserIDFromGin(c)
	ctx := c.Request.Context()

	conversations, err := s.aiDAO.FindConversationsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	return gin.H{"conversations": conversations}, nil
}

func (s *Server) handleSendAIMessage(c *gin.Context) (interface{}, error) {
	ctx := c.Request.Context()
	userID := middleware.GetUserIDFromGin(c)

	var req struct {
		BotID   int64  `json:"bot_id"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.Message == "" {
		return nil, fmt.Errorf("message is required")
	}

	// Get bot personality
	var personality *ai.Personality
	personalityKey := ""

	// Check if it's a default bot
	if req.BotID >= 9000 && req.BotID <= 9999 {
		personalities := ai.DefaultPersonalities()
		switch req.BotID {
		case 9001:
			personalityKey = "assistant"
		case 9002:
			personalityKey = "companion"
		case 9003:
			personalityKey = "tutor"
		case 9004:
			personalityKey = "creative"
		}
		personality = personalities[personalityKey]
	} else {
		// Get custom bot
		bot, err := s.aiDAO.FindBotByID(ctx, req.BotID)
		if err != nil || bot == nil {
			return nil, fmt.Errorf("bot not found")
		}
		parsed, err := ai.ParsePersonality(bot.Personality)
		if err != nil {
			return nil, fmt.Errorf("invalid personality config: %w", err)
		}
		personality = parsed
	}

	if personality == nil {
		return nil, fmt.Errorf("bot personality not found")
	}

	// Get conversation context
	convCtx := s.aiContextMgr.GetContext(req.BotID, userID)

	// Get message history (last 10 messages for context)
	history := convCtx.GetRecentMessages(10)

	// Call AI service
	response, err := s.aiBotManager.Chat(ctx, req.BotID, history, req.Message)
	if err != nil {
		log.Errorf("Failed to call AI service: %v", err)
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Add user message and AI response to context
	convCtx.AddMessage("user", req.Message)
	convCtx.AddMessage("assistant", response)

	// Save messages to database
	// User message
	userMsg := &model.Message{
		FromUserID:       userID,
		ConversationID:   req.BotID,
		ConversationType: model.ConversationTypeAI,
		MsgType:          model.MsgTypeText,
		Content:          req.Message,
		Seq:              0, // Will be set by DAO
	}
	seq, err := s.messageDAO.GetNextSeq(ctx, req.BotID, model.ConversationTypeAI)
	if err == nil {
		userMsg.Seq = seq
	}
	if err := s.messageDAO.CreateMessage(ctx, userMsg); err != nil {
		log.Errorf("Failed to save user AI message: %v", err)
	}

	// AI response message (use bot ID as from_user_id)
	aiMsg := &model.Message{
		FromUserID:       req.BotID,
		ConversationID:   req.BotID,
		ConversationType: model.ConversationTypeAI,
		MsgType:          model.MsgTypeText,
		Content:          response,
		Seq:              0,
	}
	seq, err = s.messageDAO.GetNextSeq(ctx, req.BotID, model.ConversationTypeAI)
	if err == nil {
		aiMsg.Seq = seq
	}
	if err := s.messageDAO.CreateMessage(ctx, aiMsg); err != nil {
		log.Errorf("Failed to save AI response message: %v", err)
	}

	return gin.H{
		"reply":       response,
		"bot_id":      req.BotID,
		"user_msg_id": userMsg.MsgID,
		"ai_msg_id":   aiMsg.MsgID,
	}, nil
}
