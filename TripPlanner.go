package main

import (
	"encoding/json"
	"fmt"
	//"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"mongo-tools/vendor/src/gopkg.in/mgo.v2"
	//"mongo-tools/vendor/src/gopkg.in/mgo.v2/bson"
	"httprouter"
	//"mgo.v2"
	//"mgo.v2/bson"
	"net/http"
	"strings"
	"time"
)

type LocationReq struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type LocationRes struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Address    string        `json:"address"`
	City       string        `json:"city"`
	State      string        `json:"state"`
	Zip        string        `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

type GeoLoc struct {
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

type GoogleLocationRes struct {
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

var collection *mgo.Collection
var locRes LocationRes

const (
	timeout = time.Duration(time.Second * 100)
)

func connectMongo() {
	//uri := "mongodb://arpit9691:Arpit#9691@ds043694.mongolab.com:43694/location_db"
	uri := "mongodb://nipun:nipun@ds045464.mongolab.com:45464/db2"
	ses, err := mgo.Dial(uri)

	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
	} else {
		ses.SetSafe(&mgo.Safe{})
		//collection = ses.DB("location_db").C("test")
		collection = ses.DB("db2").C("qwerty")
	}
}

func getGoogleLoc(address string) (geoLocation GeoLoc) {

	client := http.Client{Timeout: timeout}
	url := fmt.Sprintf("http://maps.google.com/maps/api/geocode/json?address=%s", address)
	res, err := client.Get(url)
	if err != nil {
		fmt.Errorf("Can't read Google API: %v", err)
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&geoLocation)
	if err != nil {
		fmt.Errorf("Error in decoding the Google: %v", err)
	}
	return geoLocation
}

func getLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id := bson.ObjectIdHex(p.ByName("locationID"))
	err := collection.FindId(id).One(&locRes)
	if err != nil {
		fmt.Printf("error finding a doc %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(200)
	json.NewEncoder(rw).Encode(locRes)
}

func addLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var tempLocReq LocationReq
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&tempLocReq)
	if err != nil {
		fmt.Errorf("Error in decoding the Input: %v", err)
	}
	address := tempLocReq.Address + " " + tempLocReq.City + " " + tempLocReq.State + " " + tempLocReq.Zip
	address = strings.Replace(address, " ", "%20", -1)

	locationDetails := getGoogleLoc(address)

	locRes.ID = bson.NewObjectId()
	locRes.Address = tempLocReq.Address
	locRes.City = tempLocReq.City
	locRes.Name = tempLocReq.Name
	locRes.State = tempLocReq.State
	locRes.Zip = tempLocReq.Zip
	locRes.Coordinate.Lat = locationDetails.Results[0].Geometry.Location.Lat
	locRes.Coordinate.Lng = locationDetails.Results[0].Geometry.Location.Lng

	err = collection.Insert(locRes)
	if err != nil {
		fmt.Printf("Can't insert document: %v\n", err)
	}

	err = collection.FindId(locRes.ID).One(&locRes)
	if err != nil {
		fmt.Printf("error finding a doc %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(201)
	json.NewEncoder(rw).Encode(locRes)
}

func updateLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var tempLocRes LocationRes
	var locRes LocationRes
	id := bson.ObjectIdHex(p.ByName("locationID"))
	err := collection.FindId(id).One(&locRes)
	if err != nil {
		fmt.Printf("error finding a doc %v\n")
	}
	tempLocRes.Name = locRes.Name
	tempLocRes.Address = locRes.Address
	tempLocRes.City = locRes.City
	tempLocRes.State = locRes.State
	tempLocRes.Zip = locRes.Zip
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&tempLocRes)

	if err != nil {
		fmt.Errorf("Error in decoding the Input: %v", err)
	}

	address := tempLocRes.Address + " " + tempLocRes.City + " " + tempLocRes.State + " " + tempLocRes.Zip
	address = strings.Replace(address, " ", "%20", -1)
	locationDetails := getGoogleLoc(address)
	tempLocRes.Coordinate.Lat = locationDetails.Results[0].Geometry.Location.Lat
	tempLocRes.Coordinate.Lng = locationDetails.Results[0].Geometry.Location.Lng
	err = collection.UpdateId(id, tempLocRes)
	if err != nil {
		fmt.Printf("got an error updating a doc %v\n")
	}

	err = collection.FindId(id).One(&locRes)
	if err != nil {
		fmt.Printf("got an error finding a doc %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(201)
	json.NewEncoder(rw).Encode(locRes)
}

func deleteLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id := bson.ObjectIdHex(p.ByName("locationID"))
	err := collection.RemoveId(id)
	if err != nil {
		fmt.Printf("got an error deleting a doc %v\n")
	}
	rw.WriteHeader(200)
}

func main() {
	mux := httprouter.New()
	mux.GET("/locations/:locationID", getLocation)
	mux.POST("/locations", addLocation)
	mux.PUT("/locations/:locationID", updateLocation)
	mux.DELETE("/locations/:locationID", deleteLocation)
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	connectMongo()
	server.ListenAndServe()
}
