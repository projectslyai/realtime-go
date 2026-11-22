# Courseexe Realtime Golang SDK

Official Go client and server publisher for **Courseexe Realtime Channels** (`https://courseexe.com`).

---

## Installation

```bash
go get github.com/projectslyai/realtime-go
```

---

## Usage

```go
package main

import (
	"log"
	realtime "github.com/projectslyai/realtime-go"
)

func main() {
	client := realtime.NewClient(realtime.Config{
		AppID:  "YOUR_APP_ID",
		Key:    "YOUR_APP_KEY",
		Secret: "YOUR_APP_SECRET",
		Host:   "courseexe.com",
		Port:   443,
		Secure: true,
	})

	err := client.Trigger("orders", "order-created", map[string]interface{}{
		"order_id": 1042,
		"status":   "dispatched",
	})
	if err != nil {
		log.Fatalf("Failed to trigger event: %v", err)
	}
	log.Println("Event triggered successfully!")
}
```

---

## License
MIT © [Courseexe](https://courseexe.com) & [PROJECTSLY AI SOLUTIONS](https://projectsly.in)
