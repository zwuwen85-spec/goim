package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Terry-Mao/goim/internal/chatapi/model"
)

// GroupDAO handles group-related database operations
type GroupDAO struct {
	mysql *MySQL
}

// NewGroupDAO creates a new GroupDAO
func NewGroupDAO(mysql *MySQL) *GroupDAO {
	return &GroupDAO{mysql: mysql}
}

// CreateGroup creates a new group
func (d *GroupDAO) CreateGroup(ctx context.Context, group *model.Group) error {
	// Generate group number if not provided
	if group.GroupNo == "" {
		group.GroupNo = generateGroupNo()
	}
	query := `INSERT INTO ` + "`groups`" + ` (group_no, name, avatar_url, owner_id, max_members, join_type, mute_all)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := d.mysql.Exec(ctx, query,
		group.GroupNo, group.Name, group.AvatarURL, group.OwnerID,
		group.MaxMembers, group.JoinType, group.MuteAll,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	group.ID = id
	return nil
}

// FindByID finds a group by ID
func (d *GroupDAO) FindByID(ctx context.Context, id int64) (*model.Group, error) {
	query := `SELECT id, group_no, name, avatar_url, owner_id, max_members, join_type, mute_all, created_at, updated_at
		FROM ` + "`groups`" + ` WHERE id = ?`
	return d.scanGroup(d.mysql.QueryRow(ctx, query, id))
}

// FindByGroupNo finds a group by group number
func (d *GroupDAO) FindByGroupNo(ctx context.Context, groupNo string) (*model.Group, error) {
	query := `SELECT id, group_no, name, avatar_url, owner_id, max_members, join_type, mute_all, created_at, updated_at
		FROM ` + "`groups`" + ` WHERE group_no = ?`
	return d.scanGroup(d.mysql.QueryRow(ctx, query, groupNo))
}

// GetGroupsByUser gets all groups for a user
func (d *GroupDAO) GetGroupsByUser(ctx context.Context, userID int64) ([]*model.GroupMember, error) {
	query := `SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.nickname, gm.joined_at,
		       g.id, g.group_no, g.name, g.avatar_url, g.owner_id, g.max_members
		FROM group_members gm
		INNER JOIN ` + "`groups`" + ` g ON gm.group_id = g.id
		WHERE gm.user_id = ?
		ORDER BY gm.joined_at DESC`
	rows, err := d.mysql.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.GroupMember
	for rows.Next() {
		member, err := d.scanGroupMemberWithGroup(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

// GetMembers gets all members of a group
func (d *GroupDAO) GetMembers(ctx context.Context, groupID int64) ([]*model.GroupMember, error) {
	query := `
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.nickname, gm.joined_at,
		       u.id, u.username, u.nickname, u.avatar_url, u.status
		FROM group_members gm
		INNER JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ?
		ORDER BY gm.role DESC, gm.joined_at ASC
	`
	rows, err := d.mysql.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.GroupMember
	for rows.Next() {
		member, err := d.scanGroupMemberWithUser(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

// AddMember adds a member to a group
func (d *GroupDAO) AddMember(ctx context.Context, groupID, userID int64, role int8) error {
	query := `
		INSERT INTO group_members (group_id, user_id, role, joined_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := d.mysql.Exec(ctx, query, groupID, userID, role, time.Now())
	return err
}

// RemoveMember removes a member from a group
func (d *GroupDAO) RemoveMember(ctx context.Context, groupID, userID int64) error {
	query := `DELETE FROM group_members WHERE group_id = ? AND user_id = ?`
	_, err := d.mysql.Exec(ctx, query, groupID, userID)
	return err
}

// UpdateMemberRole updates a member's role in the group
func (d *GroupDAO) UpdateMemberRole(ctx context.Context, groupID, userID int64, role int8) error {
	query := `UPDATE group_members SET role = ? WHERE group_id = ? AND user_id = ?`
	_, err := d.mysql.Exec(ctx, query, role, groupID, userID)
	return err
}

// UpdateMemberNickname updates a member's nickname in the group
func (d *GroupDAO) UpdateMemberNickname(ctx context.Context, groupID, userID int64, nickname string) error {
	query := `UPDATE group_members SET nickname = ? WHERE group_id = ? AND user_id = ?`
	_, err := d.mysql.Exec(ctx, query, nickname, groupID, userID)
	return err
}

// SetMemberMute sets/unsets mute for a member
func (d *GroupDAO) SetMemberMute(ctx context.Context, groupID, userID int64, muteUntil *time.Time) error {
	query := `UPDATE group_members SET mute_until = ? WHERE group_id = ? AND user_id = ?`
	_, err := d.mysql.Exec(ctx, query, muteUntil, groupID, userID)
	return err
}

// GetMemberRole gets a member's role in a group
func (d *GroupDAO) GetMemberRole(ctx context.Context, groupID, userID int64) (int8, error) {
	query := `SELECT role FROM group_members WHERE group_id = ? AND user_id = ?`
	var role int8
	err := d.mysql.QueryRow(ctx, query, groupID, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return role, err
}

// DeleteGroup deletes a group (by owner)
func (d *GroupDAO) DeleteGroup(ctx context.Context, groupID int64) error {
	// First delete all members
	_, err := d.mysql.Exec(ctx, `DELETE FROM group_members WHERE group_id = ?`, groupID)
	if err != nil {
		return err
	}
	// Then delete the group
	_, err = d.mysql.Exec(ctx, `DELETE FROM `+"`groups`"+` WHERE id = ?`, groupID)
	return err
}

// TransferOwnership transfers group ownership to another member
func (d *GroupDAO) TransferOwnership(ctx context.Context, groupID, newOwnerID int64) error {
	// Start transaction
	// Update new owner to owner role
	_, err := d.mysql.Exec(ctx, `UPDATE group_members SET role = 3 WHERE group_id = ? AND user_id = ?`, groupID, newOwnerID)
	if err != nil {
		return err
	}
	// Update old owner to admin
	_, err = d.mysql.Exec(ctx, `
		UPDATE group_members SET role = 2
		WHERE group_id = ? AND user_id = (SELECT owner_id FROM `+"`groups`"+` WHERE id = ?)
	`, groupID, groupID)
	if err != nil {
		return err
	}
	// Update group owner
	_, err = d.mysql.Exec(ctx, `UPDATE `+"`groups`"+` SET owner_id = ? WHERE id = ?`, newOwnerID, groupID)
	return err
}

// UpdateGroup updates group information
func (d *GroupDAO) UpdateGroup(ctx context.Context, group *model.Group) error {
	query := `UPDATE ` + "`groups`" + ` SET name = ?, avatar_url = ?, max_members = ?, join_type = ?, mute_all = ? WHERE id = ?`
	_, err := d.mysql.Exec(ctx, query,
		group.Name, group.AvatarURL, group.MaxMembers, group.JoinType, group.MuteAll, group.ID,
	)
	return err
}

// IsMember checks if a user is a member of a group
func (d *GroupDAO) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	query := `SELECT id FROM group_members WHERE group_id = ? AND user_id = ?`
	var id int64
	err := d.mysql.QueryRow(ctx, query, groupID, userID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateJoinRequest creates a group join request
func (d *GroupDAO) CreateJoinRequest(ctx context.Context, req *model.GroupJoinRequest) error {
	query := `
		INSERT INTO group_join_requests (group_id, user_id, message, status)
		VALUES (?, ?, ?, 1)
	`
	_, err := d.mysql.Exec(ctx, query, req.GroupID, req.UserID, req.Message)
	return err
}

// GetJoinRequests gets pending join requests for a group
func (d *GroupDAO) GetJoinRequests(ctx context.Context, groupID int64) ([]*model.GroupJoinRequest, error) {
	query := `
		SELECT id, group_id, user_id, message, status, created_at
		FROM group_join_requests
		WHERE group_id = ? AND status = 1
		ORDER BY created_at ASC
	`
	rows, err := d.mysql.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.GroupJoinRequest
	for rows.Next() {
		req, err := d.scanJoinRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

// UpdateJoinRequestStatus updates the status of a join request
func (d *GroupDAO) UpdateJoinRequestStatus(ctx context.Context, id int64, status int8) error {
	query := `UPDATE group_join_requests SET status = ? WHERE id = ?`
	_, err := d.mysql.Exec(ctx, query, status, id)
	return err
}

// scanGroup scans a group from a row
func (d *GroupDAO) scanGroup(scanner interface{ Scan(...interface{}) error }) (*model.Group, error) {
	var group model.Group
	err := scanner.Scan(
		&group.ID, &group.GroupNo, &group.Name, &group.AvatarURL,
		&group.OwnerID, &group.MaxMembers, &group.JoinType, &group.MuteAll,
		&group.CreatedAt, &group.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// scanGroupMemberWithUser scans a group member with user info
func (d *GroupDAO) scanGroupMemberWithUser(rows *sql.Rows) (*model.GroupMember, error) {
	var member model.GroupMember
	var user model.User

	err := rows.Scan(
		&member.ID, &member.GroupID, &member.UserID, &member.Role,
		&member.Nickname, &member.JoinedAt,
		&user.ID, &user.Username, &user.Nickname, &user.AvatarURL, &user.Status,
	)
	if err != nil {
		return nil, err
	}

	member.User = &user
	return &member, nil
}

// scanGroupMemberWithGroup scans a group member with group info
func (d *GroupDAO) scanGroupMemberWithGroup(rows *sql.Rows) (*model.GroupMember, error) {
	var member model.GroupMember
	var group model.Group

	err := rows.Scan(
		&member.ID, &member.GroupID, &member.UserID, &member.Role,
		&member.Nickname, &member.JoinedAt,
		&group.ID, &group.GroupNo, &group.Name, &group.AvatarURL,
		&group.OwnerID, &group.MaxMembers,
	)
	if err != nil {
		return nil, err
	}

	member.Group = &group
	return &member, nil
}

// scanJoinRequest scans a join request from a row
func (d *GroupDAO) scanJoinRequest(scanner interface{ Scan(...interface{}) error }) (*model.GroupJoinRequest, error) {
	var req model.GroupJoinRequest
	err := scanner.Scan(
		&req.ID, &req.GroupID, &req.UserID, &req.Message, &req.Status, &req.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// generateGroupNo generates a unique group number
func generateGroupNo() string {
	// Simple format: G + timestamp (e.g., G1735661200)
	return fmt.Sprintf("G%d", time.Now().Unix())
}
