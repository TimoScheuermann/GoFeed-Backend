# GoFeed

## Topics

* Was macht die Anwendung
* Hintergrundwissen
* Ein einfaches Beispiel
* Wie funktioniert die Anwendung
* Finale Bewertung von Go als Sprache für Webservices

---

## Was macht die Anwendung

Dieses Projekt ist eine Art "Playground", um Webservices mit Hilfe von Go zu realisieren. Hierfür wurde ein einfacher Feed entwickelt. Der Fokus lag hier weniger auf dem Frontend (Vue) und mehr auf dem Backend (Go).

Nutzer können sich mittels oAuth (Google & GitHub) anmelden und Beiträge verfassen, die von anderen gelesen werden können. Eigene Beiträge können bearbeitet und gelöscht werden. Hierdurch deckt die Anwendung typische Request an eine REST-Schnittstelle ab. GET, POST, DELETE und PATCH. Durch die oAuth-Integration ist auch ein kleiner Teil der Authentisierung, Authentifizierung und Autorisierung abgedeckt.

---

## Hintergrundwissen

Go ist für mich eine komplett neue Sprache gewesen. Ich hatte zuvor keine Erfahrungen mit C oder anderen Go-ähnlichen Sprachen. Mit meinem Hintergrund als Webentwickler war ich mit Sprachen wie PHP, JavaScript und Typescript sehr vertraut. Neben diesen Sprachen beherrschte ich auch Java.

Nun ging es also los mit Go. Ich nutzte hierzu "[A Tour of Go](https://tour.golang.org/welcome/1)". Dies ist eine interaktive Tour, bei der Go-Code im Browser geschrieben und getestet werden kann. Es ist eine Tutorial-Reihe, welche Schritt für Schritt den Syntax von Go erklärt.

Die Tour ist in folgende [Themenblöcke](https://tour.golang.org/list) aufgeteilt:

* Basics
  * Packages
  * Imports
  * Exported names
  * Functions
  * Multiple results
  * Named return values
  * Variables
  * Basic types
* Flow Control
  * For
  * If (else)
  * Switch
  * Defer
* More types
  * Pointers
  * Structs
  * Arrays
  * Map
* Methods
  * Erros
  * Readers
  * Images
* Concurrency
  * Goroutines
  * Channels
  * Range and Close
  * Select

Nachdem ich einschließlich das Kapitel Flow-Control durchgearbeitet hatte, fühlte ich mich bereits sicher genug meine erste Anwendung zu schreiben. Durch meine eigenen Erfahrungen wusste ich auch, dass ich es besser lerne, wenn ich es selbst schreibe und nicht halbe Lösungen vorgelegt bekomme. Gerade die Problemlösung hilft mir enorm eine neue Sprache schneller und besser zu verstehen als mich durch 20 Bücher zu arbeiten.

---

## Ein einfaches Beispiel

Dieses Beispiel zeigt, wie einfach es ist, eine Datenbank (MongoDB) mit Go zu verknüpfen und Anfragen über eine REST-Schnittstelle zu verarbeiten.

Zunächst muss Go [installiert](https://golang.org/doc/install) werden.

Ist dies erledigt kann die Entwicklung auch schon starten.

Wir erstellen ein Verzeichnis, in welchem wir unsere Anwendung schreiben möchten.

```bash
cd %HOMEPATH%

mkdir go-rest

cd go-rest
```

Als nächstes initialisieren wir unser Projekt und geben ihm einen Namen.

```bash
go mod init go-rest
```

Als nächstes erstellen wir unsere erste Go-Datei, in welcher wir unsreren Go-Code schreiben. Beispielsweise *app.go*.

Hier kann nun zunächst folgender Code eingefügt werden:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hallo zusammen!")
}
```

Jede Datei beginnt **immer** mit "package \<packagename>". Java Entwickler kennen das Konzept von Packages. Hierdurch können wir den Code in Funktionsbereiche aufteilen und gegenseitig importieren. In Code ist dies ebenfalls möglich. In dem oben gezeigten Code importieren wir sogar das fmt-Paket von Go. Jede Anwendung verfügt über ein package main und eine Funktion main. Hier startet Go auch die Anwendung.

Um den Code auszuführen, kann folgender Befehl ausgeführt werden:

```bash
# Wir befinden uns im gleichen Verzeichnis wie die Dateien go.mod und app.go

go run .

Hallo zusammen!
```

Webentwickler von node.js Anwendung sind vertraut mit NPM. Go bietet eine ähnliche Möglichkeit externe Pakete zu installieren oder eigene zu veröffentlichen. Mit dem Befehl *go get XY* können Pakete installiert werden. Das Pendant zu NPM für Go ist [pkg.go.dev](pkg.go.dev).

Für unser kleines Beispiel benötigen wir ein zwei zusätzliche Pakete, um uns die Arbeit zu erleichtern.

```bash
go get github.com/gorilla/mux
go get go.mongodb.org/mongo-driver/mongo
```

Gorilla/Mux ist ein Router, um REST-Anfragen einfach zu verarbeiten, Mongo-Driver/mongo ist unser Driver für die Datenbank Anbindung.

Beginnen wir also zunöchst mit dem Verbindungsaufbau zur Datenbank.
Hier zu erstellen wir eine init Funktion, welche bei Programmstart autom. ausgeführt wird.

```go
var database *mongo.Database

func init() {
	clientOptions := options.Client().ApplyURI("MongoDB Verbindungs URL")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	c, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to MongoDB")

	database = c.Database("go-rest")
}
```

Wir speichern die Database ab, damit wir sie später bei Anfragen wiederverwenden können.

Im nächsten Schritt kümmern wir uns um das Routing. Hierzu haben wir das Gorilla/Mux Paket installiert. Wir möchten zunächst einfache Anfragen wie POST und GET realisieren. Hierzu erstellen wir zunächst ein struct (Objekt/Klasse), in welchem wir die Datenstruktur vorgeben.

```go
type Message struct {
	MessageID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Message   string             `json:"message" bson:"message"`
}
```

Structs in go sind sehr vielfälltig, einerseits definieren sie Attribute und deren Typen, andererseits können auch Zusätze mitgegeben werden. In unserem Beispiel haben wir jeweils ein json und bson Zusatz hinzugefügt. Da wir eine Webanwendung entwickeln und die Daten als json zurückgeben werden, kann mit dem json-Zusatz definiert werden, wie die Variable umbenannt werden soll, sobald sie nach außen geschickt wird, bzw. wie sie heißen muss, wenn sie von außen kommt.
*Im Vgl. hierzu ist der bson-Zusatz für MongoDb und das dortige Mapping der Namen*.

Das omitempty gibt lediglich an, dass das Attribut auch weggelassen werden kann, wenn der Wert nil ist. Neben bson und json gibt es auch Zusatzpakete welche von diesem Syntax gebrauch machen. So kann ein validator hinzugefügt werden, welche die einzelnen Attribute validiert. Mehr zum Thema Validierung kann [hier](https://github.com/go-playground/validator) gefunden werden.

Nun bauen wir unsere Endpunkte ein

```go
func main() {
	fmt.Println("Hallo zusammen!")

	router := mux.NewRouter()
	router.HandleFunc("/", postMessage).Methods("POST")
	router.HandleFunc("/{id}", getMessage).Methods("GET")
	router.HandleFunc("/", getMessages).Methods("GET")

	log.Fatal(http.ListenAndServe(":3000", router))
}
```

Im obigen Beispiel erstellen wir zunächst unseren Router. Das defer führt den Code erst am Ende des Code-Blocks aus. Die HandleFunc-Methode nimmt 2 Parameter entgegen, zum einen den Pfad zum anderen eine Funktion, welche bei einem Aufruf ausgeführt wird.

Im nächsten Schritt müssen wir nun unsere Methoden schreiben.

```go
func postMessage(w http.ResponseWriter, req *http.Request) {
	var message Message
	json.NewDecoder(req.Body).Decode(&message)

	result, _ := database.Collection("messages").InsertOne(context.Background(), bson.M{
		"message": message.Message,
	})

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(result)
}
```

Wir definieren zunächst eine Message Variable in welches wir den Request Body parsen wollen. Im folgenden decoden wir den Body und parsen ihn in unsere Variable. Als nächstes speichern wie die Nachricht in MongoDB und geben das Ergebnis an den Aufrufer zurück.

---

```go
func getMessage(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	oid, err := primitive.ObjectIDFromHex(params["id"])

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(`{"error": "Ungültige ObjectID"}`)
		return
	}

	var message Message
	err = database.Collection("messages").FindOne(context.Background(), bson.M{"_id": oid}).Decode(&message)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(`{"error": "Message not found"}`)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(message)
}
```

Wir lesen zunächst die ID aus unseren Paremtern. Im Anschluss konvertieren wir diese zu einer ObjectID, hierbei wird auch gleichzeitig geprüft, ob es sich um eine ObjectID handelt. Falls nicht gibt es einen Error und wir geben dies an den Aufrufer zurück.

Als nächstes erstellen wir wieder unsere Messsage Variable, führen den FindOne Befehl auf der Datenbank aus, filtern nach der id und dekodieren das Ergebnis in unsere Message Variable. Sollte es hierbei zu einem Fehler kommen geben wir diesen zurück, ansonsten die Nachricht.

---

```go
func getMessages(w http.ResponseWriter, req *http.Request) {
	cursor, err := database.Collection("messages").Find(context.Background(), bson.M{})

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(`{"error": "Ein Fehler beim Aufrufen ist aufgetreten"}`)
		return
	}

	defer cursor.Close(context.Background())

	messages := []Message{}
	for cursor.Next(context.Background()) {
		var message Message
		cursor.Decode(&message)
		messages = append(messages, message)
	}

	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(`{"error": "Ein Fehler beim Iterieren ist aufgetreten"}`)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
```

Für das Auslesen mehrerer Datensätze in MongoDB wird ein cursor benötigt. Wir lesen alle Datensätze aus der Datenbank ohne zu Filtern (einfaches *bson.M{}*). Sollte bereits hier ein Fehler auftreten, informieren wir den Aufrufer. Bevor wir vergessen den cursor am Ende zu schließen, schließen wir ihn direkt nach Erstellung mittels defer.

Als nächstes definieren wir wieder unsere Messages variable, diesmal ist sie allerdings ein leeres Array. Folglich iterieren wir durch die Ergebnisse des Cursors, dekodieren die einzelne Nachricht und fügen diese dem Array hinzu. Sollte es hierbei zu einem Fehler kommen informieren wir den Client.

---

Unser kompletter Code sollte nun so aussehen:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var database *mongo.Database

type Message struct {
	MessageID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Message   string             `json:"message" bson:"message"`
}

func init() {
	clientOptions := options.Client().ApplyURI("MongoDB Verbindungs URL")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	c, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to MongoDB")

	database = c.Database("go-rest")
}

func main() {
	fmt.Println("Hallo zusammen!")

	router := mux.NewRouter()
	router.HandleFunc("/", postMessage).Methods("POST")
	router.HandleFunc("/{id}", getMessage).Methods("GET")
	router.HandleFunc("/", getMessages).Methods("GET")

	log.Fatal(http.ListenAndServe(":3000", router))
}

func postMessage(w http.ResponseWriter, req *http.Request) {
	var message Message
	json.NewDecoder(req.Body).Decode(&message)

	result, _ := database.Collection("messages").InsertOne(context.Background(), bson.M{
		"message": message.Message,
	})

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getMessage(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	oid, err := primitive.ObjectIDFromHex(params["id"])

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(`{"error": "Ungültige ObjectID"}`)
		return
	}

	var message Message
	err = database.Collection("messages").FindOne(context.Background(), bson.M{"_id": oid}).Decode(&message)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(`{"error": "Message not found"}`)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func getMessages(w http.ResponseWriter, req *http.Request) {
	cursor, err := database.Collection("messages").Find(context.Background(), bson.M{})

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(`{"error": "Ein Fehler beim Aufrufen ist aufgetreten"}`)
		return
	}

	defer cursor.Close(context.Background())

	messages := []Message{}
	for cursor.Next(context.Background()) {
		var message Message
		cursor.Decode(&message)
		messages = append(messages, message)
	}

	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(`{"error": "Ein Fehler beim Iterieren ist aufgetreten"}`)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
```

Wenn wir diesen nun mit
```bash
go run .
```
ausführen und zu [http://localhost:3000](http://localhost:3000) navigieren, sollten wir ein leeres Array als Antwort erhalten, da wir noch keine Daten in der Datenbank haben.

----

## Wie funktioniert die Anwendung?

Im Prinzip ist sie ähnlich aufgebaut wie in unserem Beispiel. Allerdings wurde der besseren Lesbarkeit auf eine einzelne Go-Datei verzichtet und dafür einzelne Pakete für die Datenbank, Nachrichten und Authentisierung erstellt.

Das Nachrichten Paket hat hier noch zusätzliche Validierer und Filter optionen wie limit und skip bei GET All Anfragen.

## TODOS:

- warum
- wie
- was so nice
- eigene erfahrungen (letztes WE)
- Code kommentieren ^^
