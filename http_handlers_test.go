package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/tinkerbell/tink/protos/packet"
)

func TestGetMetadataCacher(t *testing.T) {
	for name, test := range cacherMetadataTests {
		t.Log(name)
		hegelServer.hardwareClient = hardwareGetterMock{test.json}

		os.Setenv("DATA_MODEL_VERSION", "")

		req, err := http.NewRequest("GET", "/metadata", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.RemoteAddr = test.remote
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(getMetadata)

		handler.ServeHTTP(resp, req)

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		hw := exportedHardwareCacher{}
		err = json.Unmarshal(resp.Body.Bytes(), &hw)
		if err != nil {
			t.Error("Error in unmarshalling hardware")
		}

		if hw.ID != test.id {
			t.Errorf("handler returned unexpected id: got %v want %v",
				hw.ID, test.id)
		}
		if hw.Arch != test.arch {
			t.Errorf("handler returned unexpected arch: got %v want %v",
				hw.Arch, test.arch)
		}
		if hw.State != test.state {
			t.Errorf("handler returned unexpected state: got %v want %v",
				hw.State, test.state)
		}
		if hw.EFIBoot != test.efiBoot {
			t.Errorf("handler returned unexpected efi boot: got %v want %v",
				hw.EFIBoot, test.efiBoot)
		}
		if hw.PlanSlug != test.planSlug {
			t.Errorf("handler returned unexpected plan slug: got %v want %v",
				hw.PlanSlug, test.planSlug)
		}
		if hw.Facility != test.facility {
			t.Errorf("handler returned unexpected facility: got %v want %v",
				hw.Facility, test.facility)
		}
		if hw.Hostname != test.hostname {
			t.Errorf("handler returned unexpected hostname: got %v want %v",
				hw.Hostname, test.hostname)
		}
		if hw.BondingMode != test.bondingMode {
			t.Errorf("handler returned unexpected bonding mode: got %v want %v",
				hw.BondingMode, test.bondingMode)
		}
	}
}

func TestGetMetadataTinkerbell(t *testing.T) {
	os.Setenv("DATA_MODEL_VERSION", "1")

	for name, test := range tinkerbellMetadataTests {
		t.Log(name)
		hegelServer.hardwareClient = hardwareGetterMock{test.json}

		req, err := http.NewRequest("GET", "/metadata", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.RemoteAddr = test.remote
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(getMetadata)

		handler.ServeHTTP(resp, req)

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		hw := exportedHardwareTinkerbell{}
		err = json.Unmarshal(resp.Body.Bytes(), &hw)
		if err != nil {
			t.Error("Error in unmarshalling hardware")
		}

		if hw.ID != test.id {
			t.Errorf("handler returned unexpected id: got %v want %v",
				hw.ID, test.id)
		}

		if hw.Metadata == nil {
			return
		}

		var metadata packet.Metadata
		md, err := json.Marshal(hw.Metadata)
		if err != nil {
			t.Error("Error in marshalling hardware metadata", err)
		}
		err = json.Unmarshal(md, &metadata)
		if err != nil {
			t.Error("Error in unmarshalling hardware metadata", err)
		}

		if metadata.BondingMode != test.bondingMode {
			t.Errorf("handler returned unexpected bonding mode: got %v want %v",
				metadata.BondingMode, test.bondingMode)
		}
	}
}

func TestGetUserDataCacher(t *testing.T) {
	for name, test := range cacherUserDataTests {
		t.Log(name)
		hegelServer.hardwareClient = hardwareGetterMock{test.json}

		os.Setenv("DATA_MODEL_VERSION", "")

		req, err := http.NewRequest("GET", "/userdata", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.RemoteAddr = test.remote
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(getUserData)

		handler.ServeHTTP(resp, req)

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		if resp.Body.String() != test.userdata {
			t.Errorf("handler returned unexpected userdata: got %v want %v",
				resp.Body.String(), test.userdata)
		}

	}
}

func TestGetUserDataTinkerbell(t *testing.T) {
	os.Setenv("DATA_MODEL_VERSION", "1")

	for name, test := range tinkerbellUserDataTests {
		t.Log(name)
		hegelServer.hardwareClient = hardwareGetterMock{test.json}

		req, err := http.NewRequest("GET", "/userdata", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.RemoteAddr = test.remote
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(getUserData)

		handler.ServeHTTP(resp, req)

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		if resp.Body.String() != test.userdata {
			t.Errorf("handler returned unexpected userdata: got %v want %v",
				resp.Body.String(), test.userdata)
		}
	}
}

var cacherMetadataTests = map[string]struct {
	id       string
	remote   string
	arch string
	state string
	efiBoot bool
	planSlug string
	facility string
	hostname string
	bondingMode int
	json     string
}{
	"cacher": {
		id:       "8978e7d4-1a55-4845-8a66-a5259236b104",
		remote:   "192.168.1.5",
		arch: "x86_64",
		state: "provisioning",
		efiBoot: false,
		planSlug: "t1.small.x86",
		facility: "onprem",

		json:     cacherDataModel,
	},
}

var tinkerbellMetadataTests = map[string]struct {
	id          string
	remote      string
	bondingMode int64
	json        string
}{
	"tinkerbell": {
		id:          "fde7c87c-d154-447e-9fce-7eb7bdec90c0",
		remote:      "192.168.1.5",
		bondingMode: 5,
		json:        tinkerbellDataModel,
	},
	"tinkerbell no metadata": {
		id:     "363115b0-f03d-4ce5-9a15-5514193d131a",
		remote: "192.168.1.5",
		json:   tinkerbellNoMetadata,
	},
}

var cacherUserDataTests = map[string]struct {
	remote   string
	userdata string
	json     string
}{
	"cacher userdata": {
		remote: "192.168.1.5",
		userdata: `#!/bin/bash

echo "Hello world!"`,
		json: cacherUserData,
	},
	"cacher no userdata": {
		remote: "192.168.1.5",
		json:   cacherNoUserData,
	},
}

var tinkerbellUserDataTests = map[string]struct {
	remote   string
	userdata string
	json     string
}{
	"tinkerbell userdata": {
		remote: "192.168.1.5",
		userdata: `#!/bin/bash
echo "Hello world!"`,
		json: tinkerbellUserData,
	},
	"tinkerbell no userdata": {
		remote: "192.168.1.5",
		json:   tinkerbellNoUserData,
	}, "tinkerbell no metadata": {
		remote: "192.168.1.5",
		json:   tinkerbellNoMetadata,
	},
}