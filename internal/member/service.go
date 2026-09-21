package member

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type Service interface {
	FindAll(ctx context.Context) ([]Member, error)
	FindByID(ctx context.Context, id int64) (Member, error)
	Create(ctx context.Context, member CreateMemberRequest) (Member, error)
	Update(ctx context.Context, id int64, member UpdateMemberRequest) (Member, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo  Repository
	cache redisCache.Cache
}

func NewService(repo Repository, cache redisCache.Cache) *service {
	return &service{repo: repo, cache: cache}
}

func (s *service) FindAll(ctx context.Context) ([]Member, error) {
	cacheKey := "member:all"

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var members []Member
		if err := json.Unmarshal(val, &members); err != nil {
			log.Printf("failed to unmarshal members data from cache: %v", err)
		} else {
			return members, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch members data from cache: %v", err)
	}

	members, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(members) > 0 {
		membersBytes, err := json.Marshal(members)
		if err != nil {
			log.Printf("failed to marshal members data: %v", err)
		} else {
			if err := s.cache.Set(ctx, cacheKey, membersBytes, 10*time.Minute); err != nil {
				log.Printf("failed to set members data to cache: %v", err)
			}
		}
	}

	return members, nil
}

func (s *service) FindByID(ctx context.Context, id int64) (Member, error) {
	cacheKey := fmt.Sprintf("member:%d", id)

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var member Member
		if err := json.Unmarshal(val, &member); err != nil {
			log.Printf("failed to unmarshal member data from cache: %v", err)
		} else {
			return member, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch member data from cache: %v", err)
	}

	member, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Member{}, err
	}

	memberBytes, err := json.Marshal(member)
	if err != nil {
		log.Printf("failed to marshal member data: %v", err)
	} else {
		if err := s.cache.Set(ctx, cacheKey, memberBytes, 15*time.Minute); err != nil {
			log.Printf("failed to set member data to cache: %v", err)
		}
	}

	return member, nil
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateMemberCode(length int) (string, error) {
	result := make([]byte, length)
	max := big.NewInt(int64(len(charset)))

	for i := range length {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func (s *service) Create(ctx context.Context, member CreateMemberRequest) (Member, error) {
	memberCode, err := generateMemberCode(7)
	if err != nil {
		return Member{}, err
	}

	memberRequest := Member{
		MemberCode: memberCode,
		Name:       member.Name,
		Email:      member.Email,
		Phone:      member.Phone,
		Address:    member.Address,
	}

	newMember, err := s.repo.Create(ctx, memberRequest)
	if err != nil {
		return Member{}, err
	}

	if err := s.cache.Delete(ctx, "member:all"); err != nil {
		log.Printf("failed to invalidate member cache: %v", err)
	}

	return newMember, nil
}

func (s *service) Update(ctx context.Context, id int64, member UpdateMemberRequest) (Member, error) {
	memberUpdateRequest := Member{
		Name:    member.Name,
		Email:   member.Email,
		Phone:   member.Phone,
		Address: member.Address,
	}

	updatedMember, err := s.repo.Update(ctx, id, memberUpdateRequest)
	if err != nil {
		return Member{}, err
	}

	cacheKey := fmt.Sprintf("member:%d", id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("failed to invalidate cache %s: %v", cacheKey, err)
	}

	return updatedMember, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("member:%d", id)
	if err := s.cache.Delete(ctx, cacheKey, "member:all"); err != nil {
		log.Printf("failed to invalidate caches %s and member:all: %v", cacheKey, err)
	}

	return nil
}

func IsValidMemberCode(code string) bool {
	return len(code) == 7
}
