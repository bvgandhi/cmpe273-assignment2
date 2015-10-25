package main

//import "gopkg.in/mgo.v2/bson"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/julienschmidt/httprouter"
)

//LocnStruct locnstruct
type LocnStruct struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type seqstruct struct {
	//Greeting string `json:"greeting"`
	ID  string `json:"_id"`
	Seq int    `json:"seq"`
}

type postrespstruct struct {
	//Greeting string `json:"greeting"`
	MyID       int    `json:"id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        int    `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

type updatereqt struct {
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}
type reqtstruct struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

func postLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	log.Printf("Entered hello")
	var s LocnStruct
	t := reqtstruct{}
	var buffer bytes.Buffer
	//left part of url
	err := json.NewDecoder(req.Body).Decode(&t)
	buffer.WriteString("http://maps.google.com/maps/api/geocode/json?address=")
	add := strings.Replace(t.Address, " ", "+", -1)
	buffer.WriteString(add)
	buffer.WriteString("+")
	city := strings.Replace(t.City, " ", "+", -1)
	buffer.WriteString(city)
	buffer.WriteString("+")
	buffer.WriteString(t.State)
	buffer.WriteString("+")
	buffer.WriteString(t.Zip)
	//buffer.WriteString("+")

	//buffer.WriteString("+%s+%s+%s", t.City, t.State, t.Zip)
	buffer.WriteString("&sensor=false")
	log.Printf(buffer.String())

	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	//log.Println(t.InputString)
	resp := postrespstruct{}

	//=1600+Amphitheatre+Parkway,+Mountain+View,+CA&sensor=false"

	response, err := http.Get(buffer.String())
	if err != nil {
		fmt.Printf("error occured")
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

		json.Unmarshal([]byte(contents), &s)
		fmt.Println(s)

		// connect to mongo
		uri := "mongodb://bhumikgandhi05:b05051988@ds041404.mongolab.com:41404/bg273"

		sess, err := mgo.Dial(uri)
		if err != nil {
			fmt.Printf("Can't connect to mongo, go error %v\n", err)
			os.Exit(1)
		}
		defer sess.Close()

		sess.SetSafe(&mgo.Safe{})

		//increment  the sequence in db
		seqstructins := seqstruct{}

		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"seq": 1}},
			ReturnNew: true,
		}
		collection1 := sess.DB("bg273").C("sequence")
		_, err1 := collection1.Find(bson.M{"_id": "userid"}).Apply(change, &seqstructins)
		if err1 != nil {
			fmt.Println("got an error finding a doc")
			os.Exit(1)
		}
		fmt.Println(seqstructins)
		//end of increment logic

		//response struct to be stored in mongo and send to UI
		resp.MyID = seqstructins.Seq
		//resp.MyID = 1
		resp.Address = t.Address
		resp.City = t.City
		resp.State = t.State
		resp.Zip, err = strconv.Atoi(t.Zip)
		resp.Coordinate.Lat = s.Results[0].Geometry.Location.Lat
		resp.Coordinate.Lng = s.Results[0].Geometry.Location.Lng
		resp.Name = t.Name
		//	resp.ID = bson.NewObjectId()
		collection := sess.DB("bg273").C("resp")
		//
		//doc := postrespstruct{Id: bson.NewObjectId()}
		fmt.Println(resp)
		err = collection.Insert(resp)
		if err != nil {
			fmt.Printf("Can't insert document: %v\n", err)
			os.Exit(1)
		}

		//marshal struct to json
		repjson, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		rw.WriteHeader(201)
		fmt.Fprintf(rw, "%s", repjson)
	}
	//return nil

}

func getLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	//fmt.Fprintf(rw, "Hello, %s!\n", p.ByName("locationid"))

	// connect to mongo
	uri := "mongodb://bhumikgandhi05:b05051988@ds041404.mongolab.com:41404/bg273"

	sess, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}
	defer sess.Close()

	sess.SetSafe(&mgo.Safe{})
	collection := sess.DB("bg273").C("resp")
	response := postrespstruct{}
	fmt.Println(p.ByName("locationid"))
	intid, _ := strconv.Atoi(p.ByName("locationid"))
	err1 := collection.Find(bson.M{"myid": intid}).One(&response)
	if err1 != nil {
		panic(err1)
	}
	fmt.Println("id's query resposne", response)

	//marshal struct to json
	repjson, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	rw.WriteHeader(200)
	fmt.Fprintf(rw, "%s", repjson)
}

func deleteLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	uri := "mongodb://bhumikgandhi05:b05051988@ds041404.mongolab.com:41404/bg273"

	sess, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}
	defer sess.Close()

	sess.SetSafe(&mgo.Safe{})
	collection := sess.DB("bg273").C("resp")

	intid, _ := strconv.Atoi(p.ByName("locationid"))
	fmt.Println(p.ByName("locationid"))
	//	err1 := collection.Find(bson.M{"myid": intid}).One(&response)
	err1 := collection.Remove(bson.M{"myid": intid})
	if err1 != nil {
		panic(err1)
	} else {
		fmt.Printf("sucessfully delted %s", p.ByName("locationid"))
	}
	rw.WriteHeader(200)

}

//updatereqt

func updateLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	log.Printf("Entered updateLocation")
	reqt := updatereqt{}
	//left part of url
	err := json.NewDecoder(req.Body).Decode(&reqt)
	// connect to mongo
	uri := "mongodb://bhumikgandhi05:b05051988@ds041404.mongolab.com:41404/bg273"

	sess, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}
	defer sess.Close()

	sess.SetSafe(&mgo.Safe{})
	collection := sess.DB("bg273").C("resp")
	response := postrespstruct{}

	intid, _ := strconv.Atoi(p.ByName("locationid"))
	intzip, _ := strconv.Atoi(reqt.Zip)
	colQuerier := bson.M{"myid": intid}
	change := bson.M{"$set": bson.M{"address": reqt.Address, "city": reqt.City, "state": reqt.State, "zip": intzip}}
	err = collection.Update(colQuerier, change)
	if err != nil {
		panic(err)
	}

	err1 := collection.Find(bson.M{"myid": intid}).One(&response)
	if err1 != nil {
		panic(err1)
	}
	fmt.Println("id's query resposne", response)

	//marshal struct to json
	repjson, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	//rw.Header().Set(key string, value string)
	rw.WriteHeader(201)
	fmt.Fprintf(rw, "%s", repjson)

}

func main() {
	mux := httprouter.New()
	mux.GET("/location/:locationid", getLocation)
	mux.PUT("/location/:locationid", updateLocation)
	mux.DELETE("/location/:locationid", deleteLocation)
	mux.POST("/location", postLocation)
	server := http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
