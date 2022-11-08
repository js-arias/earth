// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package vector_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/vector"
)

func TestPixels(t *testing.T) {
	tests := map[string]struct {
		in   string
		want []int
	}{
		"south pole": {
			in: "south-pole.gpml",
			want: []int{
				38909, 39053, 39054, 39055, 39056, 39069, 39070, 39071, 39072, 39073, 39192, 39193, 39196, 39197, 39198, 39199, 39200, 39201, 39203, 39212,
				39213, 39214, 39215, 39216, 39217, 39218, 39219, 39220, 39221, 39222, 39223, 39224, 39225, 39226, 39227, 39228, 39229, 39230, 39231, 39232,
				39233, 39339, 39344, 39345, 39346, 39347, 39348, 39349, 39350, 39351, 39352, 39353, 39354, 39355, 39356, 39357, 39358, 39362, 39363, 39364,
				39365, 39366, 39367, 39368, 39369, 39370, 39371, 39372, 39373, 39374, 39375, 39376, 39377, 39378, 39379, 39380, 39381, 39382, 39383, 39384,
				39385, 39386, 39387, 39492, 39493, 39494, 39495, 39496, 39497, 39498, 39499, 39500, 39501, 39502, 39503, 39504, 39505, 39506, 39507, 39508,
				39509, 39510, 39511, 39512, 39513, 39514, 39515, 39516, 39517, 39518, 39519, 39520, 39521, 39522, 39523, 39524, 39525, 39526, 39527, 39528,
				39529, 39530, 39531, 39532, 39533, 39534, 39535, 39536, 39630, 39633, 39634, 39635, 39636, 39637, 39638, 39639, 39640, 39641, 39642, 39643,
				39644, 39645, 39646, 39647, 39648, 39649, 39650, 39651, 39652, 39653, 39654, 39655, 39656, 39657, 39658, 39659, 39660, 39661, 39662, 39663,
				39664, 39665, 39666, 39667, 39668, 39669, 39670, 39671, 39672, 39673, 39674, 39675, 39676, 39677, 39678, 39768, 39769, 39770, 39771, 39772,
				39773, 39774, 39775, 39776, 39777, 39778, 39779, 39780, 39781, 39782, 39783, 39784, 39785, 39786, 39787, 39788, 39789, 39790, 39791, 39792,
				39793, 39794, 39795, 39796, 39797, 39798, 39799, 39800, 39801, 39802, 39803, 39804, 39805, 39806, 39807, 39808, 39809, 39810, 39811, 39812,
				39813, 39814, 39898, 39899, 39900, 39901, 39902, 39903, 39904, 39905, 39906, 39907, 39908, 39909, 39910, 39911, 39912, 39913, 39914, 39915,
				39916, 39917, 39918, 39919, 39920, 39921, 39922, 39923, 39924, 39925, 39926, 39927, 39928, 39929, 39930, 39931, 39932, 39933, 39934, 39935,
				39936, 39937, 39938, 39939, 39940, 39941, 39942, 39943, 39944, 40014, 40015, 40016, 40017, 40018, 40019, 40020, 40021, 40022, 40023, 40024,
				40025, 40026, 40027, 40028, 40029, 40030, 40031, 40032, 40033, 40034, 40035, 40036, 40037, 40038, 40039, 40040, 40041, 40042, 40043, 40044,
				40045, 40046, 40047, 40048, 40049, 40050, 40051, 40052, 40053, 40054, 40055, 40056, 40057, 40058, 40059, 40060, 40061, 40062, 40063, 40064,
				40065, 40066, 40067, 40068, 40069, 40070, 40133, 40134, 40135, 40136, 40137, 40138, 40139, 40140, 40141, 40142, 40143, 40144, 40145, 40146,
				40147, 40148, 40149, 40150, 40151, 40152, 40153, 40154, 40155, 40156, 40157, 40158, 40159, 40160, 40161, 40162, 40163, 40164, 40165, 40166,
				40167, 40168, 40169, 40170, 40171, 40172, 40173, 40174, 40175, 40176, 40177, 40178, 40179, 40180, 40181, 40182, 40183, 40184, 40185, 40186,
				40187, 40188, 40248, 40249, 40250, 40251, 40252, 40253, 40254, 40255, 40256, 40257, 40258, 40259, 40260, 40261, 40262, 40263, 40264, 40265,
				40266, 40267, 40268, 40269, 40270, 40271, 40272, 40273, 40274, 40275, 40276, 40277, 40278, 40279, 40280, 40281, 40282, 40283, 40284, 40285,
				40286, 40287, 40288, 40289, 40290, 40291, 40292, 40293, 40294, 40295, 40296, 40297, 40298, 40299, 40300, 40355, 40356, 40357, 40358, 40359,
				40360, 40361, 40362, 40363, 40364, 40365, 40366, 40367, 40368, 40369, 40370, 40371, 40372, 40373, 40374, 40375, 40376, 40377, 40378, 40379,
				40380, 40381, 40382, 40383, 40384, 40385, 40386, 40387, 40388, 40389, 40390, 40391, 40392, 40393, 40394, 40395, 40396, 40397, 40398, 40399,
				40400, 40401, 40402, 40403, 40404, 40405, 40406, 40458, 40459, 40460, 40461, 40462, 40463, 40464, 40465, 40466, 40467, 40468, 40469, 40470,
				40471, 40472, 40473, 40474, 40475, 40476, 40477, 40478, 40479, 40480, 40481, 40482, 40483, 40484, 40485, 40486, 40487, 40488, 40489, 40490,
				40491, 40492, 40493, 40494, 40495, 40496, 40497, 40498, 40499, 40500, 40501, 40502, 40503, 40504, 40505, 40506, 40553, 40554, 40555, 40556,
				40557, 40558, 40559, 40560, 40561, 40562, 40563, 40564, 40565, 40566, 40567, 40568, 40569, 40570, 40571, 40572, 40573, 40574, 40575, 40576,
				40577, 40578, 40579, 40580, 40581, 40582, 40583, 40584, 40585, 40586, 40587, 40588, 40589, 40590, 40591, 40592, 40593, 40594, 40595, 40596,
				40597, 40598, 40599, 40600, 40644, 40645, 40646, 40647, 40648, 40649, 40650, 40651, 40652, 40653, 40654, 40655, 40656, 40657, 40658, 40659,
				40660, 40661, 40662, 40663, 40664, 40665, 40666, 40667, 40668, 40669, 40670, 40671, 40672, 40673, 40674, 40675, 40676, 40677, 40678, 40679,
				40680, 40681, 40682, 40683, 40684, 40685, 40686, 40727, 40728, 40729, 40730, 40731, 40732, 40733, 40734, 40735, 40736, 40737, 40738, 40739,
				40740, 40741, 40742, 40743, 40744, 40745, 40746, 40747, 40748, 40749, 40750, 40751, 40752, 40753, 40754, 40755, 40756, 40757, 40758, 40759,
				40760, 40761, 40762, 40763, 40764, 40765, 40766, 40767, 40768, 40806, 40807, 40808, 40809, 40810, 40811, 40812, 40813, 40814, 40815, 40816,
				40817, 40818, 40819, 40820, 40821, 40822, 40823, 40824, 40825, 40826, 40827, 40828, 40829, 40830, 40831, 40832, 40833, 40834, 40835, 40836,
				40837, 40838, 40839, 40840, 40841, 40842, 40877, 40878, 40879, 40880, 40881, 40882, 40883, 40884, 40885, 40886, 40887, 40888, 40889, 40890,
				40891, 40892, 40893, 40894, 40895, 40896, 40897, 40898, 40899, 40900, 40901, 40902, 40903, 40904, 40905, 40906, 40907, 40908, 40909, 40910,
				40911, 40912, 40944, 40945, 40946, 40947, 40948, 40949, 40950, 40951, 40952, 40953, 40954, 40955, 40956, 40957, 40958, 40959, 40960, 40961,
				40962, 40963, 40964, 40965, 40966, 40967, 40968, 40969, 40970, 40971, 40972, 40973, 40974, 41003, 41004, 41005, 41006, 41007, 41008, 41009,
				41010, 41011, 41012, 41013, 41014, 41015, 41016, 41017, 41018, 41019, 41020, 41021, 41022, 41023, 41024, 41025, 41026, 41027, 41028, 41029,
				41030, 41031, 41056, 41057, 41058, 41059, 41060, 41061, 41062, 41063, 41064, 41065, 41066, 41067, 41068, 41069, 41070, 41071, 41072, 41073,
				41074, 41075, 41076, 41077, 41078, 41079, 41080, 41103, 41104, 41105, 41106, 41107, 41108, 41109, 41110, 41111, 41112, 41113, 41114, 41115,
				41116, 41117, 41118, 41119, 41120, 41121, 41122, 41123, 41124, 41125, 41144, 41145, 41146, 41147, 41148, 41149, 41150, 41151, 41152, 41153,
				41154, 41155, 41156, 41157, 41158, 41159, 41160, 41161, 41162, 41178, 41179, 41180, 41181, 41182, 41183, 41184, 41185, 41186, 41187, 41188,
				41189, 41190, 41191, 41192, 41193, 41194, 41207, 41208, 41209, 41210, 41211, 41212, 41213, 41214, 41215, 41216, 41217, 41218, 41228, 41229,
				41230, 41231, 41232, 41233, 41234, 41235, 41236, 41237, 41238, 41245, 41246, 41247, 41248, 41249, 41250, 41254, 41255, 41256, 41257,
			},
		},
		"first pole": {
			in: "first-pole.gpml",
			want: []int{
				40131, 40245, 40246, 40247, 40352, 40353, 40354, 40454, 40455, 40456, 40457, 40549, 40550, 40551, 40552, 40638, 40639, 40640, 40641, 40642,
				40643, 40720, 40721, 40722, 40723, 40724, 40725, 40726, 40799, 40800, 40801, 40802, 40803, 40804, 40805, 40871, 40872, 40873, 40874, 40875,
				40876, 40940, 40941, 40942, 40943, 40990, 40997, 40998, 40999, 41000, 41001, 41002, 41045, 41046, 41049, 41050, 41051, 41052, 41053, 41054,
				41055, 41092, 41093, 41094, 41096, 41097, 41098, 41099, 41100, 41101, 41102, 41132, 41133, 41136, 41137, 41138, 41139, 41140, 41141, 41142,
				41143, 41163, 41164, 41166, 41167, 41168, 41169, 41170, 41171, 41172, 41173, 41174, 41175, 41176, 41177, 41195, 41196, 41197, 41198, 41199,
				41200, 41201, 41202, 41203, 41204, 41205, 41206, 41219, 41220, 41221, 41222, 41223, 41224, 41225, 41226, 41227, 41239, 41240, 41241, 41242,
				41243, 41244, 41251, 41252, 41253,
			},
		},
		"north pole": {
			in:   "north-pole.gpml",
			want: []int{0, 3, 4, 12, 13, 14},
		},
		"no pole": {
			in: "no-pole.gpml",
			want: []int{
				39214, 39364, 39365, 39366, 39367, 39368, 39369, 39379, 39380, 39381, 39495, 39496, 39497, 39507, 39508, 39509, 39510, 39511, 39512, 39513,
				39514, 39515, 39516, 39517, 39518, 39519, 39520, 39521, 39522, 39523, 39524, 39525, 39526, 39527, 39528, 39529, 39530, 39531, 39532, 39533,
				39534, 39535, 39630, 39635, 39636, 39637, 39638, 39639, 39640, 39641, 39642, 39643, 39644, 39645, 39646, 39647, 39648, 39649, 39650, 39651,
				39652, 39653, 39654, 39655, 39656, 39657, 39658, 39659, 39660, 39661, 39662, 39663, 39664, 39665, 39666, 39667, 39668, 39669, 39670, 39671,
				39672, 39673, 39674, 39675, 39676, 39677, 39678, 39760, 39768, 39771, 39772, 39773, 39774, 39775, 39776, 39777, 39778, 39779, 39780, 39781,
				39782, 39783, 39784, 39785, 39786, 39787, 39788, 39789, 39790, 39791, 39792, 39793, 39794, 39795, 39796, 39797, 39798, 39799, 39800, 39801,
				39802, 39803, 39804, 39805, 39806, 39807, 39808, 39809, 39810, 39811, 39812, 39813, 39814, 39891, 39898, 39899, 39900, 39901, 39902, 39903,
				39904, 39905, 39906, 39907, 39908, 39909, 39910, 39911, 39912, 39913, 39914, 39915, 39916, 39917, 39918, 39919, 39920, 39921, 39922, 39923,
				39924, 39925, 39926, 39927, 39928, 39929, 39930, 39931, 39932, 39933, 39934, 39935, 39936, 39937, 39938, 39939, 39940, 39941, 39942, 39943,
				39944, 40013, 40014, 40015, 40016, 40017, 40018, 40019, 40020, 40021, 40022, 40023, 40024, 40025, 40026, 40027, 40028, 40029, 40030, 40031,
				40032, 40033, 40034, 40035, 40036, 40037, 40038, 40039, 40040, 40041, 40042, 40043, 40044, 40045, 40046, 40047, 40048, 40049, 40050, 40051,
				40052, 40053, 40054, 40055, 40056, 40057, 40058, 40059, 40060, 40061, 40062, 40063, 40064, 40065, 40066, 40067, 40068, 40069, 40070, 40129,
				40130, 40131, 40132, 40133, 40134, 40135, 40136, 40137, 40138, 40139, 40140, 40141, 40142, 40143, 40144, 40145, 40146, 40147, 40148, 40149,
				40150, 40151, 40152, 40153, 40154, 40155, 40156, 40157, 40158, 40159, 40160, 40161, 40162, 40163, 40164, 40165, 40166, 40167, 40168, 40169,
				40170, 40171, 40172, 40173, 40174, 40175, 40176, 40177, 40178, 40179, 40180, 40181, 40182, 40183, 40184, 40185, 40186, 40187, 40188, 40243,
				40244, 40245, 40246, 40247, 40248, 40249, 40250, 40251, 40252, 40253, 40254, 40255, 40256, 40257, 40258, 40259, 40260, 40261, 40262, 40263,
				40264, 40265, 40266, 40267, 40268, 40269, 40270, 40271, 40272, 40273, 40274, 40275, 40276, 40277, 40278, 40279, 40280, 40281, 40282, 40283,
				40284, 40285, 40286, 40287, 40288, 40289, 40290, 40291, 40292, 40293, 40294, 40295, 40296, 40297, 40298, 40299, 40300, 40301, 40350, 40351,
				40352, 40353, 40354, 40355, 40356, 40357, 40358, 40359, 40360, 40361, 40362, 40363, 40364, 40365, 40366, 40367, 40368, 40369, 40370, 40371,
				40372, 40373, 40374, 40375, 40376, 40377, 40378, 40379, 40380, 40381, 40382, 40383, 40384, 40385, 40386, 40387, 40388, 40389, 40390, 40391,
				40392, 40393, 40394, 40395, 40396, 40397, 40398, 40399, 40400, 40401, 40402, 40403, 40404, 40405, 40406, 40407, 40408, 40451, 40452, 40453,
				40454, 40455, 40456, 40457, 40458, 40459, 40460, 40461, 40462, 40463, 40464, 40465, 40466, 40467, 40468, 40469, 40470, 40471, 40472, 40473,
				40474, 40475, 40476, 40477, 40478, 40479, 40480, 40481, 40482, 40483, 40484, 40485, 40486, 40487, 40488, 40489, 40490, 40491, 40492, 40493,
				40494, 40495, 40496, 40497, 40498, 40499, 40500, 40501, 40502, 40503, 40504, 40505, 40506, 40507, 40546, 40547, 40548, 40549, 40550, 40551,
				40552, 40553, 40554, 40555, 40556, 40557, 40558, 40559, 40560, 40561, 40562, 40563, 40564, 40565, 40566, 40567, 40568, 40569, 40570, 40571,
				40572, 40573, 40574, 40575, 40576, 40577, 40578, 40579, 40580, 40581, 40582, 40583, 40584, 40585, 40586, 40587, 40588, 40589, 40590, 40591,
				40592, 40593, 40594, 40595, 40596, 40597, 40598, 40599, 40600, 40636, 40637, 40638, 40639, 40640, 40641, 40642, 40643, 40644, 40645, 40646,
				40647, 40648, 40649, 40650, 40651, 40652, 40653, 40654, 40655, 40656, 40657, 40658, 40659, 40660, 40661, 40662, 40663, 40664, 40665, 40666,
				40667, 40668, 40669, 40670, 40671, 40672, 40673, 40674, 40675, 40676, 40677, 40678, 40679, 40680, 40681, 40682, 40683, 40684, 40685, 40686,
				40687, 40720, 40721, 40722, 40723, 40724, 40725, 40726, 40727, 40728, 40729, 40730, 40731, 40732, 40733, 40734, 40735, 40736, 40737, 40738,
				40739, 40740, 40741, 40742, 40743, 40744, 40745, 40746, 40747, 40748, 40749, 40750, 40751, 40752, 40753, 40754, 40755, 40756, 40757, 40758,
				40759, 40760, 40761, 40762, 40763, 40764, 40765, 40766, 40767, 40768, 40798, 40799, 40800, 40801, 40802, 40803, 40804, 40805, 40806, 40807,
				40808, 40809, 40810, 40811, 40812, 40813, 40814, 40815, 40816, 40817, 40818, 40819, 40820, 40821, 40822, 40823, 40824, 40825, 40826, 40827,
				40828, 40829, 40830, 40831, 40832, 40833, 40834, 40835, 40836, 40837, 40838, 40839, 40840, 40841, 40842, 40870, 40871, 40872, 40873, 40874,
				40875, 40876, 40877, 40878, 40879, 40880, 40881, 40882, 40883, 40884, 40885, 40886, 40887, 40888, 40889, 40890, 40891, 40892, 40893, 40894,
				40895, 40896, 40897, 40898, 40899, 40900, 40901, 40902, 40903, 40904, 40905, 40906, 40907, 40908, 40909, 40910, 40911, 40912, 40937, 40938,
				40939, 40940, 40941, 40942, 40943, 40944, 40945, 40946, 40947, 40948, 40949, 40950, 40951, 40952, 40953, 40954, 40955, 40956, 40957, 40958,
				40959, 40960, 40961, 40962, 40963, 40964, 40965, 40966, 40967, 40968, 40969, 40970, 40971, 40972, 40973, 40974, 40975, 40997, 40998, 40999,
				41000, 41001, 41002, 41003, 41004, 41005, 41006, 41007, 41008, 41009, 41010, 41011, 41012, 41013, 41014, 41015, 41016, 41017, 41018, 41019,
				41020, 41021, 41022, 41023, 41024, 41025, 41026, 41027, 41028, 41029, 41030, 41031, 41032, 41050, 41051, 41052, 41053, 41054, 41055, 41056,
				41057, 41058, 41059, 41060, 41061, 41062, 41063, 41064, 41065, 41066, 41067, 41068, 41069, 41070, 41071, 41072, 41073, 41074, 41075, 41076,
				41077, 41078, 41079, 41080, 41081, 41082, 41083, 41084, 41085, 41096, 41097, 41098, 41099, 41100, 41101, 41102, 41103, 41104, 41105, 41106,
				41107, 41108, 41109, 41110, 41111, 41112, 41113, 41114, 41115, 41116, 41117, 41118, 41119, 41120, 41121, 41122, 41123, 41124, 41125, 41126,
				41127, 41128, 41129, 41137, 41138, 41139, 41140, 41141, 41142, 41143, 41144, 41145, 41146, 41147, 41148, 41149, 41150, 41151, 41152, 41153,
				41154, 41155, 41156, 41157, 41158, 41159, 41160, 41161, 41162, 41163, 41164, 41165, 41166, 41167, 41168, 41169, 41170, 41171, 41172, 41173,
				41174, 41175, 41176, 41177, 41178, 41179, 41180, 41181, 41182, 41183, 41184, 41185, 41186, 41187, 41188, 41189, 41190, 41191, 41192, 41193,
				41194, 41195, 41196, 41197, 41198, 41199, 41200, 41201, 41202, 41203, 41204, 41205, 41206, 41207, 41208, 41209, 41210, 41211, 41212, 41213,
				41214, 41215, 41216, 41217, 41218, 41219, 41220, 41221, 41222, 41223, 41224, 41225, 41226, 41227, 41228, 41229, 41230, 41231, 41232, 41233,
				41234, 41235, 41236, 41237, 41238, 41239, 41240, 41241, 41242, 41243, 41244, 41245, 41246, 41247, 41248, 41249, 41250, 41251, 41252, 41253,
				41254, 41255, 41256, 41257,
			},
		},
		"straight line": {
			in: "straight-line.gpml",
			want: []int{
				39367, 39368, 39514, 39515, 39516, 39517, 39518, 39525, 39526, 39527, 39528, 39529, 39637, 39638, 39639, 39640, 39648, 39649, 39650, 39651,
				39652, 39653, 39654, 39655, 39656, 39657, 39658, 39659, 39660, 39661, 39662, 39663, 39664, 39665, 39666, 39667, 39668, 39669, 39670, 39671,
				39672, 39673, 39674, 39772, 39773, 39774, 39775, 39776, 39777, 39778, 39779, 39780, 39781, 39782, 39783, 39784, 39785, 39786, 39787, 39788,
				39789, 39790, 39791, 39792, 39793, 39794, 39795, 39796, 39797, 39798, 39799, 39800, 39801, 39802, 39803, 39804, 39805, 39806, 39807, 39808,
				39809, 39810, 39811, 39812, 39903, 39904, 39905, 39906, 39907, 39908, 39909, 39910, 39911, 39912, 39913, 39914, 39915, 39916, 39917, 39918,
				39919, 39920, 39921, 39922, 39923, 39924, 39925, 39926, 39927, 39928, 39929, 39930, 39931, 39932, 39933, 39934, 39935, 39936, 39937, 39938,
				39939, 39940, 39941, 39942, 40029, 40030, 40031, 40032, 40033, 40034, 40035, 40036, 40037, 40038, 40039, 40040, 40041, 40042, 40043, 40044,
				40045, 40046, 40047, 40048, 40049, 40050, 40051, 40052, 40053, 40054, 40055, 40056, 40057, 40058, 40059, 40060, 40061, 40062, 40063, 40064,
				40065, 40066, 40067, 40068, 40148, 40149, 40150, 40151, 40152, 40153, 40154, 40155, 40156, 40157, 40158, 40159, 40160, 40161, 40162, 40163,
				40164, 40165, 40166, 40167, 40168, 40169, 40170, 40171, 40172, 40173, 40174, 40175, 40176, 40177, 40178, 40179, 40180, 40181, 40182, 40183,
				40184, 40185, 40186, 40187, 40263, 40264, 40265, 40266, 40267, 40268, 40269, 40270, 40271, 40272, 40273, 40274, 40275, 40276, 40277, 40278,
				40279, 40280, 40281, 40282, 40283, 40284, 40285, 40286, 40287, 40288, 40289, 40290, 40291, 40292, 40293, 40294, 40295, 40296, 40297, 40298,
				40299, 40370, 40371, 40372, 40373, 40374, 40375, 40376, 40377, 40378, 40379, 40380, 40381, 40382, 40383, 40384, 40385, 40386, 40387, 40388,
				40389, 40390, 40391, 40392, 40393, 40394, 40395, 40396, 40397, 40398, 40399, 40400, 40401, 40402, 40403, 40472, 40473, 40474, 40475, 40476,
				40477, 40478, 40479, 40480, 40481, 40482, 40483, 40484, 40485, 40486, 40487, 40488, 40489, 40490, 40491, 40492, 40493, 40494, 40495, 40496,
				40497, 40498, 40499, 40500, 40501, 40502, 40568, 40569, 40570, 40571, 40572, 40573, 40574, 40575, 40576, 40577, 40578, 40579, 40580, 40581,
				40582, 40583, 40584, 40585, 40586, 40587, 40588, 40589, 40590, 40591, 40592, 40593, 40594, 40658, 40659, 40660, 40661, 40662, 40663, 40664,
				40665, 40666, 40667, 40668, 40669, 40670, 40671, 40672, 40673, 40674, 40675, 40676, 40677, 40678, 40679, 40680, 40681, 40741, 40742, 40743,
				40744, 40745, 40746, 40747, 40748, 40749, 40750, 40751, 40752, 40753, 40754, 40755, 40756, 40757, 40758, 40759, 40760, 40761, 40820, 40821,
				40822, 40823, 40824, 40825, 40826, 40827, 40828, 40829, 40830, 40831, 40832, 40833, 40834, 40835, 40836, 40892, 40893, 40894, 40895, 40896,
				40897, 40898, 40899, 40900, 40901, 40902, 40903, 40904, 40959, 40960, 40961, 40962, 40963, 40964, 40965, 40966,
			},
		},
	}

	pix := earth.NewPixelation(360)
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pixelsHelper(t, name, test.in, pix, test.want)
		})
	}
}

func pixelsHelper(t testing.TB, name, in string, pix *earth.Pixelation, want []int) {
	t.Helper()

	f, err := os.Open(filepath.Join(".", "testdata", in))
	if err != nil {
		t.Fatalf("%s: unable to open file %q: %v", name, in, err)
	}
	defer f.Close()

	coll, err := vector.DecodeGPML(f)
	if err != nil {
		t.Fatalf("%s: when reading %q: %v", name, in, err)
	}
	pixels := coll[0].Pixels(pix)
	if !reflect.DeepEqual(pixels, want) {
		t.Errorf("%s: got (%d pixels), want (%d pixels)\n", name, len(pixels), len(want))
	}
}

func TestRasterPoint(t *testing.T) {
	f := vector.Feature{
		Name:  "Erebus",
		Type:  vector.HotSpot,
		Plate: 1,
		Begin: 200_000_000,
		Point: &vector.Point{Lat: -77.99999999999999, Lon: 167.00000000000006},
	}

	pix := earth.NewPixelation(360)
	pixel := f.Pixels(pix)
	if len(pixel) != 1 {
		t.Fatalf("pixels: got %d, want %d", len(pixel), 1)
	}

	pt := earth.NewPoint(f.Point.Lat, f.Point.Lon)
	dist := earth.Distance(pt, pix.ID(pixel[0]).Point())

	if dist > 0.01 {
		t.Errorf("point: got %d, want %v [dist = %.3f]", pixel[0], pix.Pixel(f.Point.Lat, f.Point.Lon), dist)
	}
}