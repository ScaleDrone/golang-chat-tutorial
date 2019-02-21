package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

const (
	scaledroneID     = "YOUR_SCALEDRONE_ID"     // 👈 PS! Replace this with your own channel ID 🚨
	scaledroneSecret = "YOUR_SCALEDRONE_SECRET" // 👈 PS! Replace this with your own channel secret 🚨
	port             = ":8080"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/auth", auth).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static"))).Methods("GET")
	fmt.Printf("Server is running on localhost%s", port)
	panic(http.ListenAndServe(port, r))
}

type customClaims struct {
	jwt.StandardClaims
	Client      string                      `json:"client"`
	Channel     string                      `json:"channel"`
	Data        userData                    `json:"data"`
	Permissions map[string]permissionClaims `json:"permissions"`
}

type permissionClaims struct {
	Publish   bool `json:"publish"`
	Subscribe bool `json:"subscribe"`
}

type userData struct {
	Color string `json:"color"`
	Name  string `json:"name"`
}

func getRandomName() string {
	adjs := []string{"autumn", "hidden", "bitter", "misty", "silent", "empty", "dry", "dark", "summer", "icy", "delicate", "quiet", "white", "cool", "spring", "winter", "patient", "twilight", "dawn", "crimson", "wispy", "weathered", "blue", "billowing", "broken", "cold", "damp", "falling", "frosty", "green", "long", "late", "lingering", "bold", "little", "morning", "muddy", "old", "red", "rough", "still", "small", "sparkling", "throbbing", "shy", "wandering", "withered", "wild", "black", "young", "holy", "solitary", "fragrant", "aged", "snowy", "proud", "floral", "restless", "divine", "polished", "ancient", "purple", "lively", "nameless"}
	nouns := []string{"waterfall", "river", "breeze", "moon", "rain", "wind", "sea", "morning", "snow", "lake", "sunset", "pine", "shadow", "leaf", "dawn", "glitter", "forest", "hill", "cloud", "meadow", "sun", "glade", "bird", "brook", "butterfly", "bush", "dew", "dust", "field", "fire", "flower", "firefly", "feather", "grass", "haze", "mountain", "night", "pond", "darkness", "snowflake", "silence", "sound", "sky", "shape", "surf", "thunder", "violet", "water", "wildflower", "wave", "water", "resonance", "sun", "wood", "dream", "cherry", "tree", "fog", "frost", "voice", "paper", "frog", "smoke", "star"}
	return adjs[rand.Intn(len(adjs))] + "_" + nouns[rand.Intn(len(nouns))]
}

func getRandomColor() string {
	return "#" + strconv.FormatInt(rand.Int63n(0xFFFFFF), 16)
}

func auth(w http.ResponseWriter, r *http.Request) {
	clientID := r.FormValue("clientID")
	if clientID == "" {
		http.Error(w, "No clientID defined", http.StatusUnprocessableEntity)
		return
	}

	// public room
	publicRoomRegex := "^observable-room$"
	// private room of the request user
	userPrivateRoomRegex := fmt.Sprintf("^private-room-%s$", clientID)
	// private rooms of every user besides the request user
	otherUsersPrivateRoomsRegex := fmt.Sprintf("^private-room-(?!%s$).+$", clientID)
	claims := customClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 3).Unix(),
		},
		Client:  clientID,
		Channel: scaledroneID,
		Data: userData{
			Name:  getRandomName(),
			Color: getRandomColor(),
		},
		Permissions: map[string]permissionClaims{
			publicRoomRegex: permissionClaims{ // public room
				Publish:   true, // allow publishing to public chatroom
				Subscribe: true, // allow subscribing to public chatroom
			},
			userPrivateRoomRegex: permissionClaims{
				Publish:   false, // no need to publish to ourselves
				Subscribe: true,  // allow subscribing to private messages
			},
			otherUsersPrivateRoomsRegex: permissionClaims{
				Publish:   true,  // allow publishing to other users
				Subscribe: false, // don't allow subscribing to messages sent to other users
			},
		},
	}

	// Create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret
	tokenString, err := token.SignedString([]byte(scaledroneSecret))
	if err != nil {
		http.Error(w, "Unable to sign the token", http.StatusUnprocessableEntity)
		return
	}
	// Send the token to the user
	w.Write([]byte(tokenString))
}
