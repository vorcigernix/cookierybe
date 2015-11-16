package cookieryapi

//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// gosites is an App Engine JSON backend for managing a site list.
//
// It supports the following commands:
//
// - Create a new site
// POST /sites
// > {"text": "do this"}
// < {"id": 1, "text": "do this", "created": 1356724843.0, "done": false}
//
// - Update an existing site
// POST /sites
// > {"id": 1, "text": "do this", "created": 1356724843.0, "done": true}
// < {"id": 1, "text": "do this", "created": 1356724843.0, "done": true}
//
// - List existing sites:
// GET /sites
// >
// < [{"id": 1, "text": "do this", "created": 1356724843.0, "done": true},
//    {"id": 2, "text": "do that", "created": 1356724849.0, "done": false}]
//
// - Delete 'done' sites:
// DELETE /sites
// >
// <


import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
  "strconv"

	"appengine"
	"appengine/datastore"
)

func defaultSiteList(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "SiteList", "default", 0, nil)
}

type Site struct {
	Id   int64  `json:"id" datastore:"-"`
  Name string   `json:"name"`
	URL string `json:"url" datastore:",noindex"`
  CategoryID []string `json:"categoryid"`
	Created time.Time `json:"created"`
}

func (t *Site) key(c appengine.Context) *datastore.Key {
	if t.Id == 0 {
		t.Created = time.Now()
		return datastore.NewIncompleteKey(c, "Site", defaultSiteList(c))
	}
	return datastore.NewKey(c, "Site", "", t.Id, defaultSiteList(c))
}

func (t *Site) save(c appengine.Context) (*Site, error) {
	k, err := datastore.Put(c, t.key(c), t)
	if err != nil {
		return nil, err
	}
	t.Id = k.IntID()
	return t, nil
}

func decodeSite(r io.ReadCloser) (*Site, error) {
	defer r.Close()
	var site Site
	err := json.NewDecoder(r).Decode(&site)
	return &site, err
}

func getAllSites(c appengine.Context) ([]Site, error) {
	sites := []Site{}
	ks, err := datastore.NewQuery("Site").Ancestor(defaultSiteList(c)).Order("Created").GetAll(c, &sites)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(sites); i++ {
		sites[i].Id = ks[i].IntID()
	}
	return sites, nil
}

func getCategorySites(c appengine.Context, cat int) ([]Site, error) {
	sites := []Site{}
	ks, err := datastore.NewQuery("Site").Ancestor(defaultSiteList(c)).Filter("CategoryID=", strconv.Itoa(cat)).Order("Created").GetAll(c, &sites)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(sites); i++ {
		sites[i].Id = ks[i].IntID()
	}
	return sites, nil
}


func init() {
	http.HandleFunc("/sites", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	c := appengine.NewContext(r)
	val, err := handleSites(c, r)
	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}
	if err != nil {
		c.Errorf("site error: %#v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleSites(c appengine.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "POST":
		site, err := decodeSite(r.Body)
		if err != nil {
			return nil, err
		}
		return site.save(c)
	case "GET":
    {
      if cat, err := strconv.Atoi(r.URL.Query().Get("cat")); err == nil {
        return getCategorySites(c, cat)
      }
      return getAllSites(c)
    }
	}
	return nil, fmt.Errorf("method not implemented")
}
