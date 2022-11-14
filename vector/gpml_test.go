// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package vector_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/js-arias/earth/vector"
)

func TestDecodeGPML(t *testing.T) {
	want := []vector.Feature{
		{
			Name:  "Pacific",
			Type:  vector.Basin,
			Plate: 901,
			Begin: 400_000,
			Polygon: vector.Polygon{
				{Lat: 19.85355599999994, Lon: -155.08441699999997},
				{Lat: 19.729971999999975, Lon: -155.087806},
				{Lat: 19.738222000000007, Lon: -155.00502799999998},
				{Lat: 19.519139000000024, Lon: -154.80575},
				{Lat: 19.346417000000088, Lon: -154.97772200000003},
				{Lat: 19.136611000000016, Lon: -155.50566700000002},
				{Lat: 18.913056000000097, Lon: -155.67533299999997},
				{Lat: 18.998167000000024, Lon: -155.78688900000003},
				{Lat: 19.085082999999997, Lon: -155.91097199999996},
				{Lat: 19.346499999999935, Lon: -155.88933300000002},
				{Lat: 19.72963900000005, Lon: -156.06461099999996},
				{Lat: 19.98366699999994, Lon: -155.83116699999994},
				{Lat: 20.197389000000015, Lon: -155.90624999999997},
				{Lat: 20.27277799999996, Lon: -155.853389},
				{Lat: 19.975000000000023, Lon: -155.210139},
				{Lat: 19.85355599999994, Lon: -155.08441699999997},
			},
		},
		{
			Type:  vector.Boundary,
			Plate: 802,
			Begin: 11_000_000,
			End:   2_009_999,
			Polygon: vector.Polygon{
				{Lat: -80.79844450127769, Lon: -40.3582129976047},
				{Lat: -79.55516184791934, Lon: -38.470427378950134},
				{Lat: -78.83808877386726, Lon: -38.190231537803335},
				{Lat: -80.6304551537465, Lon: -39.52871253007798},
				{Lat: -80.79844450127769, Lon: -40.3582129976047},
			},
		},
		{
			Name:  "Pacific",
			Type:  vector.Coastline,
			Plate: 901,
			Begin: 400_000,
			Polygon: []vector.Point{
				{Lat: 19.85355599999994, Lon: -155.08441699999997},
				{Lat: 19.729971999999975, Lon: -155.087806},
				{Lat: 19.738222000000007, Lon: -155.00502799999998},
				{Lat: 19.975000000000023, Lon: -155.210139},
				{Lat: 19.85355599999994, Lon: -155.08441699999997},
			},
		},
		{
			Name:  "Mexico",
			Type:  vector.Continent,
			Plate: 104,
			Begin: 170_000_000,
			Polygon: []vector.Point{
				{Lat: 24.876141470067925, Lon: -90.71179050315442},
				{Lat: 24.846500000000106, Lon: -90.8983},
				{Lat: 24.670299999999997, Lon: -91.50799999999998},
				{Lat: 24.8857604004408, Lon: -90.65126645154885},
				{Lat: 24.876141470067925, Lon: -90.71179050315442},
			},
		},
		{
			Name:  "Afif Abas",
			Type:  vector.Craton,
			Plate: 5031,
			Begin: 1_100_000_000,
			Polygon: []vector.Point{
				{Lat: 14.296135569657187, Lon: 43.63036797405223},
				{Lat: 15.025321203658427, Lon: 43.778679100251466},
				{Lat: 15.403806549449442, Lon: 44.06110599048722},
				{Lat: 13.828231160033386, Lon: 43.76544002263138},
				{Lat: 14.296135569657187, Lon: 43.63036797405223},
			},
		},
		{
			Name:  "East Avalonia",
			Type:  vector.Fragment,
			Plate: 315,
			Begin: 750_000_000,
			Polygon: []vector.Point{
				{Lat: 52.368243554845435, Lon: -9.839247481759568},
				{Lat: 52.512772727272704, Lon: -9.680281818181825},
				{Lat: 52.747569189000046, Lon: -7.994038992999949},
				{Lat: 52.28666363636366, Lon: -9.835836363636334},
				{Lat: 52.368243554845435, Lon: -9.839247481759568},
			},
		},
		{
			Name:  "Mt Read; Tyennan",
			Type:  vector.Generic,
			Plate: 8013,
			Begin: 1_400_000_000,
			Polygon: []vector.Point{
				{Lat: -41.09902777777778, Lon: 146.1386138888889},
				{Lat: -41.123955, Lon: 146.18265194444444},
				{Lat: -41.13883777777778, Lon: 146.24105888888892},
				{Lat: -41.11605888888889, Lon: 146.08567194444444},
				{Lat: -41.09902777777778, Lon: 146.1386138888889},
			},
		},
		{
			Name:  "Erebus",
			Type:  vector.HotSpot,
			Plate: 1,
			Begin: 200_000_000,
			Point: &vector.Point{Lat: -77.99999999999999, Lon: 167.00000000000006},
		},
		{
			Name:  "Tonga Ridge",
			Type:  vector.IslandArc,
			Plate: 821,
			Begin: 40_000_000,
			Polygon: []vector.Point{
				{Lat: -14.55209999999994, Lon: -174.01899999999998},
				{Lat: -14.750999999999976, Lon: -173.58199999999997},
				{Lat: -14.750999999999976, Lon: -173.29999999999998},
				{Lat: -14.442831978999946, Lon: -174.13105408099997},
				{Lat: -14.55209999999994, Lon: -174.01899999999998},
			},
		},
		{
			Name:  "Alpha Ridge - Alvey et al. (2008) Fig1a bathymetry",
			Type:  vector.LIP,
			Plate: 101,
			Begin: 82_000_000,
			Polygon: []vector.Point{
				{Lat: 85.15149947364364, Lon: 180},
				{Lat: 83.70132112817831, Lon: 180},
				{Lat: 83.87057508572468, Lon: 178.17567629054798},
				{Lat: 85.08306008527849, Lon: 178.86696019920456},
				{Lat: 85.15149947364364, Lon: 180},
			},
		},
		{
			Name:  "Baltica",
			Type:  vector.PaleoBoundary,
			Plate: 330,
			Begin: 600_000_000,
			Polygon: []vector.Point{
				{Lat: 54.82363692828839, Lon: 10.777831991680983},
				{Lat: 54.84955715330746, Lon: 10.657130556533875},
				{Lat: 54.90860909090914, Lon: 10.681945454545485},
				{Lat: 54.943327272727274, Lon: 10.838890909090935},
				{Lat: 54.82363692828839, Lon: 10.777831991680983},
			},
		},
		{
			Type:  vector.Passive,
			Plate: 109,
			Begin: 600_000_000,
			Polygon: []vector.Point{
				{Lat: 35.970860373884804, Lon: -75.65686015909654},
				{Lat: 36.00711093352829, Lon: -75.63543300892727},
				{Lat: 35.801518181818224, Lon: -75.53264545454542},
				{Lat: 35.89749090909096, Lon: -75.58979090909088},
				{Lat: 35.970860373884804, Lon: -75.65686015909654},
			},
		},
		{
			Type:  vector.Suture,
			Plate: 109,
			Begin: 600_000_000,
			Polygon: []vector.Point{
				{Lat: 39.66700668165139, Lon: -75.52153777640282},
				{Lat: 40.04696165945661, Lon: -74.78679073738832},
				{Lat: 40.16952433014954, Lon: -74.49701232299299},
				{Lat: 39.62040909090912, Lon: -75.5572181818182},
				{Lat: 39.66700668165139, Lon: -75.52153777640282},
			},
		},
		{
			Name:  "Dzabkhan block",
			Type:  vector.Terrane,
			Plate: 4101,
			Begin: 600_000_000,
			Polygon: []vector.Point{
				{Lat: 48.45773718740095, Lon: 93.96755050944287},
				{Lat: 48.97437301506492, Lon: 94.18103368257763},
				{Lat: 49.377155672328584, Lon: 94.41653935386935},
				{Lat: 47.91010461561801, Lon: 93.87272945873477},
				{Lat: 48.45773718740095, Lon: 93.96755050944287},
			},
		},
	}

	f, err := os.Open(filepath.Join(".", "testdata", "plates.gpml"))
	if err != nil {
		t.Fatalf("unable to open file \"plates.gpml\": %v", err)
	}
	defer f.Close()

	coll, err := vector.DecodeGPML(f)
	if err != nil {
		t.Fatalf("while reading \"plates.gpml\": %v", err)
	}
	if len(coll) != len(want) {
		t.Errorf("invalid decoded data: got %d elements, want %d", len(coll), len(want))
	}
	for i, c := range coll {
		if !reflect.DeepEqual(c, want[i]) {
			t.Errorf("invalid decoded data: element %d\n\tgot %v\t\nwant %v", i, c, want[i])
		}
	}
}
