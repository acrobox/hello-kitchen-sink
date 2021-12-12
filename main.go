package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "incr":
			err := redisIncr()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		case "touch":
			err := postgresTouch()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown command '%s'.", os.Args[1])
		}
		return
	}
	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":8080", Handler: mux}
	mux.HandleFunc("/", index)
	go func() {
		usr1 := make(chan os.Signal, 1)
		signal.Notify(usr1, syscall.SIGUSR1)
		for {
			<-usr1
			err := postgresReset()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			err = redisReset()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
	}()
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func index(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	postgresN, err := postgresGet()
	if err == nil {
		fmt.Fprintf(w, "%d\n", postgresN)
	} else {
		fmt.Fprintf(w, "%s\n", err)
	}
	redisN, err := redisGet()
	if err == nil || err == redis.Nil {
		fmt.Fprintf(w, "%d\n", redisN)
	} else {
		fmt.Fprintf(w, "%s\n", err)
	}
}

func postgresGet() (int64, error) {
	var n int64
	ctx := context.Background()
	dsn := os.Getenv("ABX_STORE_DSN")
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return n, err
	}
	defer conn.Close(ctx)
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM data").Scan(&n)
	return n, err
}

func postgresTouch() error {
	ctx := context.Background()
	dsn := os.Getenv("ABX_STORE_DSN")
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, "INSERT INTO data (id) VALUES (default)")
	return err
}

func postgresReset() error {
	ctx := context.Background()
	dsn := os.Getenv("ABX_STORE_DSN")
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, "TRUNCATE TABLE data")
	return err
}

func redisGet() (int64, error) {
	ctx := context.Background()
	dsn := os.Getenv("ABX_CACHE_DSN")
	options, err := redis.ParseURL(dsn)
	if err != nil {
		return 0, err
	}
	rdb := redis.NewClient(options)
	defer rdb.Close()
	image := os.Getenv("ABX_IMAGE")
	return rdb.Get(ctx, image).Int64()
}

func redisIncr() error {
	ctx := context.Background()
	dsn := os.Getenv("ABX_CACHE_DSN")
	options, err := redis.ParseURL(dsn)
	if err != nil {
		return err
	}
	rdb := redis.NewClient(options)
	defer rdb.Close()
	image := os.Getenv("ABX_IMAGE")
	return rdb.Incr(ctx, image).Err()
}

func redisReset() error {
	ctx := context.Background()
	dsn := os.Getenv("ABX_CACHE_DSN")
	options, err := redis.ParseURL(dsn)
	if err != nil {
		return err
	}
	rdb := redis.NewClient(options)
	defer rdb.Close()
	image := os.Getenv("ABX_IMAGE")
	return rdb.Set(ctx, image, 0, 0).Err()
}
