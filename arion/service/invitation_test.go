package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/common/response"
	mock_database "github.com/zeirash/recapo/arion/mock/database"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
)

func Test_iservice_InviteAdmin(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		shopID    int
		userID    int
		email     string
		lang      string
		mockSetup func(ctrl *gomock.Controller)
		wantErr   bool
	}{
		{
			name:      "invalid email format returns error",
			shopID:    1,
			userID:    2,
			email:     "not-an-email",
			lang:      "en",
			mockSetup: func(ctrl *gomock.Controller) {},
			wantErr:   true,
		},
		{
			name:   "caller is not owner returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Role: "admin"}, nil)
				userStore = mockUser
			},
			wantErr: true,
		},
		{
			name:   "GetUserByID returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(nil, errors.New("db error"))
				userStore = mockUser
			},
			wantErr: true,
		},
		{
			name:   "caller not found returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(nil, nil)
				userStore = mockUser
			},
			wantErr: true,
		},
		{
			name:   "GetSubscriptionByShopID returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(nil, errors.New("db error"))
				subscriptionStore = mockSub
			},
			wantErr: true,
		},
		{
			name:   "GetPlanByID returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(nil, errors.New("db error"))
				subscriptionStore = mockSub
			},
			wantErr: true,
		},
		{
			name:   "CountUsersByShopID returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				mockUser.EXPECT().
					CountUsersByShopID(gomock.Any(), 1).
					Return(0, errors.New("db error"))
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 2}, nil)
				subscriptionStore = mockSub
			},
			wantErr: true,
		},
		{
			name:   "plan limit reached returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				mockUser.EXPECT().
					CountUsersByShopID(gomock.Any(), 1).
					Return(2, nil)
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 2}, nil)
				subscriptionStore = mockSub
			},
			wantErr: true,
		},
		{
			name:   "email already registered returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "invite@example.com").
					Return(&model.User{ID: 5, Email: "invite@example.com"}, nil)
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(nil, nil)
				subscriptionStore = mockSub
			},
			wantErr: true,
		},
		{
			name:   "pending invitation already exists returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "invite@example.com").
					Return(nil, nil)
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(nil, nil)
				subscriptionStore = mockSub

				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetPendingInvitationByEmail(gomock.Any(), 1, "invite@example.com").
					Return(&model.Invitation{ID: 10, Status: "pending"}, nil)
				invitationStore = mockInvitation
			},
			wantErr: true,
		},
		{
			name:   "shop not found returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "invite@example.com").
					Return(nil, nil)
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(nil, nil)
				subscriptionStore = mockSub

				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetPendingInvitationByEmail(gomock.Any(), 1, "invite@example.com").
					Return(nil, nil)
				invitationStore = mockInvitation

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					GetShopByID(gomock.Any(), 1).
					Return(nil, nil)
				shopStore = mockShop
			},
			wantErr: true,
		},
		{
			name:   "CreateInvitation returns error",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "invite@example.com").
					Return(nil, nil)
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(nil, nil)
				subscriptionStore = mockSub

				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetPendingInvitationByEmail(gomock.Any(), 1, "invite@example.com").
					Return(nil, nil)
				mockInvitation.EXPECT().
					CreateInvitation(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(nil, errors.New("db error"))
				invitationStore = mockInvitation

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					GetShopByID(gomock.Any(), 1).
					Return(&model.Shop{ID: 1, Name: "Test Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop
			},
			wantErr: true,
		},
		{
			name:   "successfully invite admin",
			shopID: 1,
			userID: 2,
			email:  "invite@example.com",
			lang:   "en",
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 2).
					Return(&model.User{ID: 2, Name: "Owner", Role: "owner"}, nil)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "invite@example.com").
					Return(nil, nil)
				userStore = mockUser

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 1).
					Return(nil, nil)
				subscriptionStore = mockSub

				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetPendingInvitationByEmail(gomock.Any(), 1, "invite@example.com").
					Return(nil, nil)
				mockInvitation.EXPECT().
					CreateInvitation(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(&model.Invitation{ID: 10, ShopID: 1, Email: "invite@example.com", Status: "pending"}, nil)
				invitationStore = mockInvitation

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					GetShopByID(gomock.Any(), 1).
					Return(&model.Shop{ID: 1, Name: "Test Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop

				cfg = config.Config{FrontendURL: "https://app.example.com"}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldUser, oldShop, oldInvitation, oldSub := userStore, shopStore, invitationStore, subscriptionStore
			defer func() {
				userStore = oldUser
				shopStore = oldShop
				invitationStore = oldInvitation
				subscriptionStore = oldSub
			}()

			tt.mockSetup(ctrl)

			var s iservice
			gotErr := s.InviteAdmin(context.Background(), tt.shopID, tt.userID, tt.email, tt.lang)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("InviteAdmin() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("InviteAdmin() succeeded unexpectedly")
			}
		})
	}
}

func Test_iservice_ValidateInviteToken(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		token     string
		mockSetup func(ctrl *gomock.Controller)
		want      *response.InvitationData
		wantErr   bool
	}{
		{
			name:  "invitation not found returns error",
			token: "badtoken",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "badtoken").
					Return(nil, nil)
				invitationStore = mockInvitation
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "invitation status not pending returns error",
			token: "acceptedtoken",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "acceptedtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Status: "accepted"}, nil)
				invitationStore = mockInvitation
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "GetInvitationByToken returns error",
			token: "sometoken",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "sometoken").
					Return(nil, errors.New("db error"))
				invitationStore = mockInvitation
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "shop not found returns error",
			token: "validtoken",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					GetShopByID(gomock.Any(), 5).
					Return(nil, nil)
				shopStore = mockShop
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "successfully validate invite token",
			token: "validtoken",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					GetShopByID(gomock.Any(), 5).
					Return(&model.Shop{ID: 5, Name: "My Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop
			},
			want: &response.InvitationData{
				Email:    "invite@example.com",
				ShopName: "My Shop",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldInvitation, oldShop := invitationStore, shopStore
			defer func() {
				invitationStore = oldInvitation
				shopStore = oldShop
			}()

			tt.mockSetup(ctrl)

			var s iservice
			got, gotErr := s.ValidateInviteToken(context.Background(), tt.token)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateInviteToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateInviteToken() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateInviteToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_iservice_AcceptInvite(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		token     string
		username  string
		password  string
		mockSetup func(ctrl *gomock.Controller)
		wantErr   bool
	}{
		{
			name:     "invitation not found returns error",
			token:    "badtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "badtoken").
					Return(nil, nil)
				invitationStore = mockInvitation
			},
			wantErr: true,
		},
		{
			name:     "invitation already accepted returns error",
			token:    "acceptedtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "acceptedtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "accepted", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation
			},
			wantErr: true,
		},
		{
			name:     "GetInvitationByToken returns error",
			token:    "sometoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "sometoken").
					Return(nil, errors.New("db error"))
				invitationStore = mockInvitation
			},
			wantErr: true,
		},
		{
			name:     "GetSubscriptionByShopID returns error on accept",
			token:    "validtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(nil, errors.New("db error"))
				subscriptionStore = mockSub
			},
			wantErr: true,
		},
		{
			name:     "plan limit reached on accept returns error",
			token:    "validtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(&model.Subscription{ID: 1, ShopID: 5, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 2}, nil)
				subscriptionStore = mockSub

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					CountUsersByShopID(gomock.Any(), 5).
					Return(2, nil)
				userStore = mockUser
			},
			wantErr: true,
		},
		{
			name:     "subscription not found returns error",
			token:    "validtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(nil, nil)
				subscriptionStore = mockSub
			},
			wantErr: true,
		},
		{
			name:     "weak password returns error",
			token:    "validtoken",
			username: "New Admin",
			password: "weak",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(&model.Subscription{ID: 1, ShopID: 5, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 0}, nil)
				subscriptionStore = mockSub
			},
			wantErr: true,
		},
		{
			name:     "db.Begin failure returns error",
			token:    "validtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(&model.Subscription{ID: 1, ShopID: 5, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 0}, nil)
				subscriptionStore = mockSub

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(nil, errors.New("tx error"))
				dbGetter = func() database.DB { return mockDB }
			},
			wantErr: true,
		},
		{
			name:     "CreateUser failure returns error",
			token:    "validtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(&model.Subscription{ID: 1, ShopID: 5, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 0}, nil)
				subscriptionStore = mockSub

				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "New Admin", "invite@example.com", gomock.Any(), "admin", 5).
					Return(nil, errors.New("db error"))
				userStore = mockUser
			},
			wantErr: true,
		},
		{
			name:     "tx.Commit failure returns error",
			token:    "validtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(&model.Subscription{ID: 1, ShopID: 5, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 0}, nil)
				subscriptionStore = mockSub

				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(errors.New("commit error"))
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "New Admin", "invite@example.com", gomock.Any(), "admin", 5).
					Return(&model.User{ID: 10, ShopID: 5, Name: "New Admin", Email: "invite@example.com", Role: "admin", CreatedAt: fixedTime}, nil)
				userStore = mockUser
			},
			wantErr: true,
		},
		{
			name:     "AcceptInvitation failure returns error",
			token:    "validtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				mockInvitation.EXPECT().
					AcceptInvitation(gomock.Any(), 1).
					Return(errors.New("db error"))
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(&model.Subscription{ID: 1, ShopID: 5, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 0}, nil)
				subscriptionStore = mockSub

				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "New Admin", "invite@example.com", gomock.Any(), "admin", 5).
					Return(&model.User{ID: 10, ShopID: 5, Name: "New Admin", Email: "invite@example.com", Role: "admin", CreatedAt: fixedTime}, nil)
				userStore = mockUser
			},
			wantErr: true,
		},
		{
			name:     "successfully accept invite",
			token:    "validtoken",
			username: "New Admin",
			password: "pass1234",
			mockSetup: func(ctrl *gomock.Controller) {
				mockInvitation := mock_store.NewMockInvitationStore(ctrl)
				mockInvitation.EXPECT().
					GetInvitationByToken(gomock.Any(), "validtoken").
					Return(&model.Invitation{ID: 1, ShopID: 5, Email: "invite@example.com", Status: "pending", CreatedAt: fixedTime}, nil)
				mockInvitation.EXPECT().
					AcceptInvitation(gomock.Any(), 1).
					Return(nil)
				invitationStore = mockInvitation

				mockSub := mock_store.NewMockSubscriptionStore(ctrl)
				mockSub.EXPECT().
					GetSubscriptionByShopID(gomock.Any(), 5).
					Return(&model.Subscription{ID: 1, ShopID: 5, PlanID: 2}, nil)
				mockSub.EXPECT().
					GetPlanByID(gomock.Any(), 2).
					Return(&model.Plan{ID: 2, MaxUsers: 0}, nil)
				subscriptionStore = mockSub

				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				newUser := &model.User{ID: 10, ShopID: 5, Name: "New Admin", Email: "invite@example.com", Role: "admin", CreatedAt: fixedTime}

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "New Admin", "invite@example.com", gomock.Any(), "admin", 5).
					Return(newUser, nil)
				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 10, gomock.Any()).
					Return(nil)
				userStore = mockUser

				mockToken := mock_store.NewMockTokenStore(ctrl)
				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), gomock.Any(), gomock.Any(), 2).
					Return("access-token", nil)
				mockToken.EXPECT().
					CreateRefreshToken(gomock.Any(), gomock.Any(), gomock.Any(), 168).
					Return("refresh-token", nil)
				tokenStore = mockToken

				cfg = config.Config{SecretKey: "testsecret"}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldInvitation, oldUser, oldToken, oldSub := invitationStore, userStore, tokenStore, subscriptionStore
			oldDBGetter := dbGetter
			defer func() {
				invitationStore = oldInvitation
				userStore = oldUser
				tokenStore = oldToken
				subscriptionStore = oldSub
				dbGetter = oldDBGetter
			}()

			tt.mockSetup(ctrl)

			var s iservice
			got, gotErr := s.AcceptInvite(context.Background(), tt.token, tt.username, tt.password)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AcceptInvite() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AcceptInvite() succeeded unexpectedly")
			}

			if got.AccessToken == "" || got.RefreshToken == "" {
				t.Errorf("AcceptInvite() returned empty tokens: %+v", got)
			}
		})
	}
}
