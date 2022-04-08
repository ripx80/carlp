package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ripx80/carlp/pkgs/parser"
)

type GetTest struct {
	desc string
	str  string
	json string
}

var getTests = []GetTest{
	{
		desc: `key value pair`,
		str:  `a=b`,
		json: `{"a":"b"}`,
	},
	{
		desc: `key string pair`,
		str:  `version="Libra v3.3.2"`,
		json: `{"version":"Libra v3.3.2"}`,
	},
	{
		desc: `key digit pair`,
		str:  `version_control_revision=86054`,
		json: `{"version_control_revision":86054}`,
	},
	{
		desc: `key negative digit pair`,
		str:  `numbernegative=-34243`,
		json: `{"numbernegative":-34243}`,
	},
	{
		desc: `key float pair`,
		str:  `floatthing=1.20348`,
		json: `{"floatthing":1.2034800052642822}`,
	},
	{
		desc: `key float pair`,
		str:  `floatnegative=-1.20348`,
		json: `{"floatnegative":-1.2034800052642822}`,
	},
	{
		desc: `key nested strings`,
		str: `required_dlcs={
			"Anniversary Portraits"
			"Apocalypse"
			"Federations"
			"Horizon Signal"
			"Leviathans Story Pack"
			"Synthetic Dawn Story Pack"
			"Utopia"
		}`,
		json: `{"required_dlcs":["Anniversary Portraits","Apocalypse","Federations","Horizon Signal","Leviathans Story Pack","Synthetic Dawn Story Pack","Utopia"]}`,
	},
	{
		desc: `key nested arrays without keys`,
		str: `player=    {
			{
				name="user1"
				country     =0
			}
			{
				name="user2"
				country=1
			}

		}`,
		json: `{"player":[{"country":0,"name":"user1"},{"country":1,"name":"user2"}]}`,
	},
	{
		desc: `nested without any key`,
		str: `{
			nkey=happens
		}
		`,
		// that is a problem! no key in root: set uknown
		json: `{"unkown":{"nkey":"happens"}}`,
	},
	{
		desc: `coordinate pairs`,
		str: `coordinate={
			x=121.1325
			y=-31.49625
			origin=182
		}`,
		json: `{"coordinate":{"origin":182,"x":121.13249969482422,"y":-31.49625015258789}}`,
	},
	{
		desc: `key digit array`,
		str: `spy_networks={
			52 56 221 218 16777453 16777452 16777479 16777376 50331792
		}`,
		json: `{"spy_networks":[52,56,221,218,16777453,16777452,16777479,16777376,50331792]}`,
	},
	{
		desc: `key digit array without newlines`,
		str:  `random={ 0 4049908188 }`,
		json: `{"random":[0,4049908188]}`,
	},
	{
		desc: `key nested without equal sign inside`,
		str: `intel_manager={
			intel={
					{
					1 {
						intel=70
						stale_intel={
						}
					}
				}
			}
		}`,
		json: `{"intel_manager":{"intel":[{"1":{"intel":70,"stale_intel":{}}}]}}`,
	},
	{
		desc: `key array with no escape chars and digit dots in key`,
		str: `flags={
			custom_start_screen=62808000
			tutorial_level_picked=62808000
			anomaly_outcome_happened_anomaly.650=62832096
			anomaly_outcome_happened_anomaly.630=62838648
			Story7=62899104
			}`,
		json: `{"flags":{"Story7":62899104,"anomaly_outcome_happened_anomaly.630":62838648,"anomaly_outcome_happened_anomaly.650":62832096,"custom_start_screen":62808000,"tutorial_level_picked":62808000}}`,
	},
	{
		desc: `artefacts unicode chars`,
		str: `					effect="Kleinere Artefakte gefunden:
		Yminor_artifacts|1 1.00!
		"`,
		json: `{"effect":"Kleinere Artefakte gefunden:\n\t\t\u0011Y\u0013minor_artifacts|1 1.00\u0011!\n\t\t"}`,
	},
	{
		desc: `nested array keyval oneline`,
		str: `starbase_mgr={
			starbases={
				0={
					level="starbase_level_starhold"
					modules={
						0=shipyard				1=trading_hub			}
					buildings={
						0=hydroponics_bay			}
					build_queue=603
					shipyard_build_queue=604
					ship_design=29
					station=0
					owner=0
					orbitals={
						0=4294967295
						1=4294967295
						2=4294967295
					}
				}
			}
		}`,
		json: `{"starbase_mgr":{"starbases":{"0":{"build_queue":603,"buildings":{"0":"hydroponics_bay"},"level":"starbase_level_starhold","modules":{"0":"shipyard","1":"trading_hub"},"orbitals":{"0":4294967295,"1":4294967295,"2":4294967295},"owner":0,"ship_design":29,"shipyard_build_queue":604,"station":0}}}}`,
	},
	{
		desc: `duplicate key in root`,
		str: `nebula={
			coordinate={
				x=-217.11
				y=28.37
				origin=4294967295
				randomized=yes
			}
			name="Rebenthi Dust Clouds"
			radius=30
			galactic_object=29
			galactic_object=75
			galactic_object=92
			galactic_object=285
		}
		nebula={
			coordinate={
				x=16.9
				y=-234.92
				origin=4294967295
				randomized=yes
			}
			name="Tyjanock Expanse"
			radius=30
			galactic_object=140
			galactic_object=259
			galactic_object=324
			galactic_object=335
			galactic_object=346
		}`,
		json: `{"nebula":[{"coordinate":{"origin":4294967295,"randomized":"yes","x":-217.11000061035156,"y":28.3700008392334},"galactic_object":[285,92,29,75],"name":"Rebenthi Dust Clouds","radius":30},{"coordinate":{"origin":4294967295,"randomized":"yes","x":16.899999618530273,"y":-234.9199981689453},"galactic_object":[346,335,324,140,259],"name":"Tyjanock Expanse","radius":30}]}`,
	},
	{
		desc: `duplicate key in nested array`,
		str: `coordinate={
			x=16.9
			y=-234.92
			origin=4294967295
			randomized=yes
		}
		coordinate={
			x=16.9
			y=-234.92
			origin=4294967295
			randomized=yes
		}`,
		json: `{"coordinate":[{"origin":4294967295,"randomized":"yes","x":16.899999618530273,"y":-234.9199981689453},{"origin":4294967295,"randomized":"yes","x":16.899999618530273,"y":-234.9199981689453}]}`,
	},
	{
		desc: `nested array in map with duplicate keys`,
		str: `nebula={
			name="Rebenthi Dust Clouds"
			radius=30
			galactic_object=29
			galactic_object=75
			galactic_object=92
			galactic_object=285
		}`,
		json: `{"nebula":{"galactic_object":[285,92,29,75],"name":"Rebenthi Dust Clouds","radius":30}}`,
	},
	{
		desc: `broken newlines`,
		str: `variables={
			unrest_50
=0
		}`,
		// this is a problem
		json: `{"variables":["unrest_50"]}`,
	},
	{
		desc: `key with empty string`,
		str:  `picture=""`,
		// this is a problem
		json: `{"picture":""}`,
	},
}

/*template
{
	desc: `nested array in oneline`,
	str: ``,
	json: ``,
},
*/

/*

check mixed array and map in nested array -> must fail with error
check scanned line numbers
check key=>array on root level overrite old ones

*/

func countRune(s string, r rune) uint64 {
	var count uint64
	for _, c := range s {
		if c == r {
			count++
		}
	}
	return count
}

func TestGet(t *testing.T) {
	for _, test := range getTests {
		fmt.Println("Running:", test.desc)

		r := strings.NewReader(test.str)
		p := parser.NewParser(r)
		data, _, err := p.Parse()
		if err != nil {
			t.Errorf("%s test, expected %s, obtained %s (err %v)", test.desc, data, test.json, err)
		}
		// lnc := countRune(test.str, '\n')
		// if ln != lnc {
		// 	t.Errorf("%s test, expected %d lines, obtained %d lines", test.desc, lnc, ln)
		// }

		b, err := json.Marshal(data)
		if err != nil {
			t.Errorf("error: %v", err)
			return
		}
		json := string(b)

		if json != test.json {
			t.Errorf("%s test, expected %s, obtained %s (err %v)", test.desc, test.json, json, err)
		}

	}
}

/*TestCompGet will broke the structure, dont run it*/
func NOTestShortFile(t *testing.T) {
	fp, err := os.Open(filepath.Join("test", "short.txt"))
	if err != nil {
		t.Error(err)
		return
	}
	defer fp.Close()

	parser := parser.NewParser(fp)
	data, _, err := parser.Parse()
	if err != nil {
		t.Error(err)
	}

	parseb, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	jsonb, err := ioutil.ReadFile(filepath.Join("test", "short.json"))
	if err != nil {
		t.Error(err)
		return
	}

	if res := bytes.Compare(parseb, jsonb); res != 0 {
		t.Errorf("parsed json and json in file are not equal")
	}

}
