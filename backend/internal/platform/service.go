package platform

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	
	// SDK Client
	"github.com/zitadel/zitadel-go/v3/pkg/client/management"
	
	// Protobuf Definitions
	managementpb "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	objectpb "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object"
	userpb "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user"
	
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	platformv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/platform/v1"
	"github.com/abdul/vortyx/backend/internal/platform/db"
)

type Service struct {
	db        *db.Queries
	zitadel   *management.Client
	zitadelProjectID string
}

func NewService(pool *pgxpool.Pool, zitadelClient *management.Client, zitadelProjectID string) *Service {
	return &Service{
		db:      db.New(pool),
		zitadel: zitadelClient,
		zitadelProjectID: zitadelProjectID,
	}
}

// CreateUser creates a user in Zitadel.
func (s *Service) CreateUser(
	ctx context.Context,
	req *connect.Request[platformv1.CreateUserRequest],
) (*connect.Response[platformv1.CreateUserResponse], error) {
	if s.zitadel == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("zitadel client not configured"))
	}

	// Create human user
	resp, err := s.zitadel.AddHumanUser(ctx, &managementpb.AddHumanUserRequest{
		UserName: req.Msg.Username,
		Profile: &managementpb.AddHumanUserRequest_Profile{
			FirstName: req.Msg.FirstName,
			LastName:  req.Msg.LastName,
			DisplayName: fmt.Sprintf("%s %s", req.Msg.FirstName, req.Msg.LastName),
		},
		Email: &managementpb.AddHumanUserRequest_Email{
			Email: req.Msg.Email,
			IsEmailVerified: true, // Auto-verify for admin created users
		},
		InitialPassword: req.Msg.Password,
	})

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create user: %w", err))
	}

	// TODO: Assign role if needed (req.Msg.Role) via Grants

	// Log audit
	s.logAudit(ctx, "user.created", "user", resp.GetUserId(), map[string]interface{}{
		"username": req.Msg.Username,
		"email":    req.Msg.Email,
	})

	return connect.NewResponse(&platformv1.CreateUserResponse{
		User: &platformv1.User{
			Id:        resp.GetUserId(),
			Username:  req.Msg.Username,
			Email:     req.Msg.Email,
			FirstName: req.Msg.FirstName,
			LastName:  req.Msg.LastName,
			CreatedAt: timestamppb.Now(),
			State:     "active", // Default
		},
	}), nil
}

// ListUsers lists users from Zitadel.
func (s *Service) ListUsers(
	ctx context.Context,
	req *connect.Request[platformv1.ListUsersRequest],
) (*connect.Response[platformv1.ListUsersResponse], error) {
	if s.zitadel == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("zitadel client not configured"))
	}

	limit := uint32(req.Msg.PageSize)
	if limit == 0 {
		limit = 10
	}
	offset := uint64((req.Msg.Page - 1) * int32(limit))
	if offset < 0 {
		offset = 0
	}

	queries := []*userpb.SearchQuery{}
	if req.Msg.SearchQuery != "" {
		queries = append(queries, &userpb.SearchQuery{
			Query: &userpb.SearchQuery_UserNameQuery{
				UserNameQuery: &userpb.UserNameQuery{
					UserName: req.Msg.SearchQuery,
					Method:   objectpb.TextQueryMethod_TEXT_QUERY_METHOD_CONTAINS,
				},
			},
		})
	}

	resp, err := s.zitadel.ListUsers(ctx, &managementpb.ListUsersRequest{
		Query: &objectpb.ListQuery{
			Limit:  limit,
			Offset: offset,
		},
		Queries: queries,
	})

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list users: %w", err))
	}

	users := make([]*platformv1.User, 0, len(resp.GetResult()))
	for _, u := range resp.GetResult() {
		users = append(users, mapZitadelUser(u))
	}

	return connect.NewResponse(&platformv1.ListUsersResponse{
		Users:      users,
		TotalCount: int32(resp.GetDetails().GetTotalResult()),
	}), nil
}

func (s *Service) GetUser(
	ctx context.Context,
	req *connect.Request[platformv1.GetUserRequest],
) (*connect.Response[platformv1.GetUserResponse], error) {
	if s.zitadel == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("zitadel client not configured"))
	}

	resp, err := s.zitadel.GetUserByID(ctx, &managementpb.GetUserByIDRequest{Id: req.Msg.UserId})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user: %w", err))
	}

	return connect.NewResponse(&platformv1.GetUserResponse{User: mapZitadelUser(resp.GetUser())}), nil
}

// DeleteUser deletes a user from Zitadel.
func (s *Service) DeleteUser(
	ctx context.Context,
	req *connect.Request[platformv1.DeleteUserRequest],
) (*connect.Response[platformv1.DeleteUserResponse], error) {
	if s.zitadel == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("zitadel client not configured"))
	}

	_, err := s.zitadel.RemoveUser(ctx, &managementpb.RemoveUserRequest{
		Id: req.Msg.UserId,
	})

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete user: %w", err))
	}

	s.logAudit(ctx, "user.deleted", "user", req.Msg.UserId, nil)

	return connect.NewResponse(&platformv1.DeleteUserResponse{
		Success: true,
	}), nil
}

func (s *Service) UpdateUserRole(
	ctx context.Context,
	req *connect.Request[platformv1.UpdateUserRoleRequest],
) (*connect.Response[platformv1.UpdateUserRoleResponse], error) {
	if s.zitadel == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("zitadel client not configured"))
	}
	if s.zitadelProjectID == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("zitadel project id not configured"))
	}
	if req.Msg.UserId == "" || req.Msg.Role == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user_id and role are required"))
	}

	grants, err := s.zitadel.ListUserGrants(ctx, &managementpb.ListUserGrantRequest{
		Query: &objectpb.ListQuery{Limit: 1, Offset: 0},
		Queries: []*userpb.UserGrantQuery{
			{Query: &userpb.UserGrantQuery_UserIdQuery{UserIdQuery: &userpb.UserGrantUserIDQuery{UserId: req.Msg.UserId}}},
			{Query: &userpb.UserGrantQuery_ProjectIdQuery{ProjectIdQuery: &userpb.UserGrantProjectIDQuery{ProjectId: s.zitadelProjectID}}},
		},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list user grants: %w", err))
	}

	if len(grants.GetResult()) > 0 {
		_, err = s.zitadel.UpdateUserGrant(ctx, &managementpb.UpdateUserGrantRequest{
			UserId:   req.Msg.UserId,
			GrantId:  grants.GetResult()[0].GetId(),
			RoleKeys: []string{req.Msg.Role},
		})
	} else {
		_, err = s.zitadel.AddUserGrant(ctx, &managementpb.AddUserGrantRequest{
			UserId:    req.Msg.UserId,
			ProjectId: s.zitadelProjectID,
			RoleKeys:  []string{req.Msg.Role},
		})
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user role: %w", err))
	}

	userResp, err := s.zitadel.GetUserByID(ctx, &managementpb.GetUserByIDRequest{Id: req.Msg.UserId})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user: %w", err))
	}

	return connect.NewResponse(&platformv1.UpdateUserRoleResponse{User: mapZitadelUser(userResp.GetUser())}), nil
}

func (s *Service) ListRoles(
	ctx context.Context,
	req *connect.Request[platformv1.ListRolesRequest],
) (*connect.Response[platformv1.ListRolesResponse], error) {
	_ = req
	if s.zitadel == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("zitadel client not configured"))
	}
	if s.zitadelProjectID == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("zitadel project id not configured"))
	}

	resp, err := s.zitadel.ListProjectRoles(ctx, &managementpb.ListProjectRolesRequest{
		ProjectId: s.zitadelProjectID,
		Query:     &objectpb.ListQuery{Limit: 100, Offset: 0},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list roles: %w", err))
	}

	roles := make([]*platformv1.Role, 0, len(resp.GetResult()))
	for _, r := range resp.GetResult() {
		roles = append(roles, &platformv1.Role{
			Key:         r.GetKey(),
			DisplayName: r.GetDisplayName(),
			Description: "",
		})
	}

	return connect.NewResponse(&platformv1.ListRolesResponse{Roles: roles}), nil
}

// GetAuditLogs retrieves audit logs from local DB.
func (s *Service) GetAuditLogs(
	ctx context.Context,
	req *connect.Request[platformv1.GetAuditLogsRequest],
) (*connect.Response[platformv1.GetAuditLogsResponse], error) {
	limit := int32(req.Msg.PageSize)
	if limit == 0 {
		limit = 10
	}
	offset := int32((req.Msg.Page - 1) * limit)
	if offset < 0 {
		offset = 0
	}

	// Use nullable text for user_id filter
	var userID pgtype.Text
	if req.Msg.UserId != "" {
		userID = pgtype.Text{String: req.Msg.UserId, Valid: true}
	}

	logs, err := s.db.ListAuditLogs(ctx, db.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
		UserID: userID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list audit logs: %w", err))
	}

	count, err := s.db.CountAuditLogs(ctx, userID)
	if err != nil {
		fmt.Printf("failed to count logs: %v\n", err)
	}

	pbLogs := make([]*platformv1.AuditLog, 0, len(logs))
	for _, l := range logs {
		var details *structpb.Struct
		// TODO: Convert JSONB to Struct
		
		pbLogs = append(pbLogs, &platformv1.AuditLog{
			Id:           l.ID.String(), // pgtype.UUID implements fmt.Stringer
			UserId:       l.UserID,
			Username:     l.Username,
			Action:       l.Action,
			ResourceType: l.ResourceType,
			ResourceId:   l.ResourceID,
			Details:      details,
			CreatedAt:    timestamppb.New(l.CreatedAt.Time),
		})
	}

	return connect.NewResponse(&platformv1.GetAuditLogsResponse{
		Logs:       pbLogs,
		TotalCount: int32(count),
	}), nil
}

func (s *Service) GetSystemStats(
	ctx context.Context,
	req *connect.Request[platformv1.GetSystemStatsRequest],
) (*connect.Response[platformv1.GetSystemStatsResponse], error) {
	_ = req
	var totalUsers int32
	if s.zitadel != nil {
		resp, err := s.zitadel.ListUsers(ctx, &managementpb.ListUsersRequest{Query: &objectpb.ListQuery{Limit: 1, Offset: 0}})
		if err == nil {
			totalUsers = int32(resp.GetDetails().GetTotalResult())
		}
	}
	return connect.NewResponse(&platformv1.GetSystemStatsResponse{TotalUsers: totalUsers}), nil
}

func mapZitadelUser(u *userpb.User) *platformv1.User {
	if u == nil {
		return nil
	}

	var email, firstName, lastName string
	if human := u.GetHuman(); human != nil {
		email = human.GetEmail().GetEmail()
		firstName = human.GetProfile().GetFirstName()
		lastName = human.GetProfile().GetLastName()
	} else if machine := u.GetMachine(); machine != nil {
		firstName = machine.GetName()
		lastName = "[Machine]"
	}

	return &platformv1.User{
		Id:        u.GetId(),
		Username:  u.GetUserName(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		State:     u.GetState().String(),
		CreatedAt: u.GetDetails().GetCreationDate(),
		UpdatedAt: u.GetDetails().GetChangeDate(),
	}
}

// Helper to log audit events
func (s *Service) logAudit(ctx context.Context, action, resType, resID string, details map[string]interface{}) {
	userID, _ := ctx.Value("user_id").(string)
	
	fmt.Printf("AUDIT: %s by %s on %s:%s\n", action, userID, resType, resID)
}
