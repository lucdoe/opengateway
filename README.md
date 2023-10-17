# Capstone Monolith to Microservices

## Notes

Run:

```bash
go run cmd/capstone/main.go
```

Build:

```bash
go build -o capstone-project cmd/capstone-project/main.go
```

env:

```go
import (
	"os"
)
dbHost := os.Getenv("DB_HOST")
```

Handling JSON:

```go
import (
    "encoding/json"
)

type User struct {
    Firstname string `json:"firstname"`
    Lastname  string `json:"lastname"`
    Age       int    `json:"age"`
}

func main() {
    http.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
        var user User
        json.NewDecoder(r.Body).Decode(&user)

        fmt.Fprintf(w, "%s %s is %d years old!", user.Firstname, user.Lastname, user.Age)
    })
}
```

Hashing:

```go
import (
    "golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

Redis Rate Limiting (+ SHA256):

```go
import (
    "github.com/go-redis/redis"
)

func rateLimiterMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr

        // Hash the IP address
        hash := sha256.New()
        hash.Write([]byte(ip))
        hashedIP := hex.EncodeToString(hash.Sum(nil))

        key := fmt.Sprintf("rate:limiter:%s", hashedIP)

        // Increment the counter
        pipe := redisClient.TxPipeline()
        pipe.Incr(ctx, key)
        pipe.Expire(ctx, key, 1*time.Minute)

        _, err := pipe.Exec(ctx)
        if err != nil {
            http.Error(w, "Server Error", http.StatusInternalServerError)
            return
        }

        count, err := redisClient.Get(ctx, key).Int()
        if err != nil {
            http.Error(w, "Server Error", http.StatusInternalServerError)
            return
        }

        // Here you can set your limit per minute
        if count > 10 {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```
