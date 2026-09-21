package main

import (
	"context"
	"ferdinand/library-management-system-api/internal/auth"
	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/book"
	"ferdinand/library-management-system-api/internal/borrowing"
	"ferdinand/library-management-system-api/internal/category"
	"ferdinand/library-management-system-api/internal/config"
	"ferdinand/library-management-system-api/internal/database"
	"ferdinand/library-management-system-api/internal/member"
	"ferdinand/library-management-system-api/internal/redis"
	"ferdinand/library-management-system-api/internal/router"
	"ferdinand/library-management-system-api/internal/user"
	"fmt"
	"log"
	"time"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}


	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()



	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		fmt.Printf("failed to connect to database:%v", err)
		return
	}

	fmt.Println("Connected to database")


	redisClientService := redis.NewRedisClient(cfg.Redis)


	_,err = redisClientService.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	fmt.Println("Connected to redis!")

	redisClientCache := redis.NewCache(redisClientService)


	// repo
	bookRepo := book.NewRepository(db)
	categoryRepo := category.NewRepository(db)
	authorRepo := author.NewRepository(db)
	memberRepo := member.NewRepository(db)
	borrowingRepo := borrowing.NewRepository(db)
	userRepo := user.NewRepository(db)

	// service
	bookService := book.NewService(bookRepo, authorRepo, categoryRepo, redisClientCache)
	categoryService := category.NewService(categoryRepo, redisClientCache)
	authorService := author.NewService(authorRepo, redisClientCache)
	memberService := member.NewService(memberRepo, redisClientCache)
	borrowingService := borrowing.NewService(db, borrowingRepo, memberRepo, bookRepo)
	userService := user.NewService(userRepo, redisClientCache)
	jwtService := auth.NewJWTService(cfg.JWT.SecretKey)
	authService := auth.NewService(userRepo,jwtService)

	// handler
	bookHandler := book.NewHandler(bookService)
	categoryHandler := category.NewHandler(categoryService)
	authorHandler := author.NewHandler(authorService)
	memberHandler := member.NewHandler(memberService)
	borrowingHandler := borrowing.NewHandler(borrowingService)
	userHandler :=user.NewHandler(userService)
	authHandler := auth.NewHandler(authService)

	handlers := router.RouteHandlers{
		BookHandler:      bookHandler,
		CategoryHandler:  categoryHandler,
		AuthorHandler:    authorHandler,
		MemberHandler:    memberHandler,
		BorrowingHandler: borrowingHandler,
		UserHandler: userHandler,
		AuthHandler: authHandler,
	}

	r := router.New(handlers,jwtService)
	err = r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}

}
