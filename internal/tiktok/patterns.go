package tiktok

import "strings"

// CountryProfile holds detection patterns for a specific country/region.
type CountryProfile struct {
	Name        string
	PhoneCodes  []string
	Cities      []string
	Keywords    []string
	Hashtags    []string
	Slang       []string
	TLDs        []string
	Currencies  []string
	MusicGenres []string
	Weight      float64 // base weight multiplier for this country
}

// Countries is the ordered list of country profiles, Jamaica-first and Caribbean-prioritised.
var Countries = []CountryProfile{
	{
		Name: "Jamaica",
		PhoneCodes: []string{
			"+1876", "+1-876", "(876)", "876-", "1876",
			"876 ", "tel:+1876",
		},
		Cities: []string{
			"kingston", "portmore", "montego bay", "montego", "spanish town",
			"ocho rios", "mandeville", "half way tree", "halfwaytree",
			"new kingston", "may pen", "savanna-la-mar", "savanna la mar",
			"port antonio", "falmouth", "lucea", "black river",
			"morant bay", "old harbour", "linstead", "browns town",
			"brown's town", "negril", "dunrobin", "constant spring",
			"waterford", "gregory park", "waterhouse", "august town",
			"barbican", "cherry gardens", "norbrook", "stony hill",
		},
		Keywords: []string{
			"jamaica", "jamaican", "jah", "876", "jamrock", "jamdown",
			"yard", "yardman", "yardwoman", "inna di yard", "born ya",
			"j.a.", "ja vibes", "jamaicanmade", "irieman", "jerk chicken",
			"jerk pork", "ackee", "bammy", "festival", "patty",
			"guinness punch", "roots wine", "rum bar", "appleton",
			"rum punch", "red stripe", "lion energy",
		},
		Hashtags: []string{
			"jamaica", "jamaican", "876", "kingstonjamaica", "kingston",
			"dancehall", "reggae", "yardie", "jamaicanvibes",
			"jamaicanstyle", "jamdown", "jamaicanmade", "reggaemusic",
			"jamaicanculture", "876life", "jamaicanfood", "jamaicanpride",
			"dancehallmusic", "montegobay", "mobayrising", "ochorios",
			"negriljamaica", "portmore", "jamaicanqueen", "jamaicanking",
		},
		Slang: []string{
			"wagwan", "wha gwaan", "whagwan",
			"mi nuh", "mi a", "mi deh",
			"yuh nuh", "yuh deh", "yuh know",
			"nuh lie", "nah lie",
			"inna", "pickney", "dutty", "bless up",
			"biggup", "big up", "gyalis", "gyallis",
			"likkle", "likkle more",
			"one love", "bredda", "bredren", "sistas",
			"jancrow", "badman", "badgyal", "sketel",
			"wutless", "link mi", "suh it go", "nuff",
			"bare gyal", "bare man", "nuff respect",
			"real talk", "ting", "dutty wine",
			"rasta", "rastafari", "zion", "babylon",
			"irie", "jah bless", "give thanks",
			"no sah", "no seh", "lawd", "lawdamercy",
			"a suh", "di", "de road", "road code",
		},
		TLDs:        []string{".jm"},
		Currencies:  []string{"jmd", "j$", "jamaican dollar", "jamaican dollars"},
		MusicGenres: []string{"dancehall", "reggae", "roots reggae", "dub", "ska", "lovers rock"},
		Weight:      1.5,
	},
	{
		Name: "Trinidad and Tobago",
		PhoneCodes: []string{"+1868", "(868)", "868-", "868 "},
		Cities: []string{
			"port of spain", "san fernando", "chaguanas", "arima",
			"point fortin", "diego martin", "tunapuna", "princes town",
			"tobago", "scarborough", "couva",
		},
		Keywords:    []string{"trinidad", "tobago", "trini", "868", "tt", "trinbago", "trindad"},
		Hashtags:    []string{"trinidad", "tobago", "trini", "868", "soca", "trinbago", "carnival", "trinidadian"},
		Slang:       []string{"allyuh", "gyul", "breds", "tabanca", "wining", "liming", "mamaguy", "dotish", "ole talk"},
		TLDs:        []string{".tt"},
		Currencies:  []string{"ttd"},
		MusicGenres: []string{"soca", "calypso", "chutney"},
		Weight:      1.2,
	},
	{
		Name: "Barbados",
		PhoneCodes: []string{"+1246", "(246)", "246-", "246 "},
		Cities:      []string{"bridgetown", "speightstown", "oistins", "holetown", "bathsheba", "christ church"},
		Keywords:    []string{"barbados", "bajan", "246", "bim", "bajans"},
		Hashtags:    []string{"barbados", "bajan", "246", "bim", "bajans", "bridgetown"},
		Slang:       []string{"wunna", "um", "lef", "hey hey", "real bad", "real good"},
		TLDs:        []string{".bb"},
		Currencies:  []string{"bbd"},
		MusicGenres: []string{"soca", "spouge", "reggae"},
		Weight:      1.2,
	},
	{
		Name: "Guyana",
		PhoneCodes: []string{"+592", "592-", "592 "},
		Cities:      []string{"georgetown", "linden", "new amsterdam", "anna regina", "bartica"},
		Keywords:    []string{"guyana", "guyanese", "592", "mudland", "caricom"},
		Hashtags:    []string{"guyana", "guyanese", "592", "georgetown"},
		Slang:       []string{"alyuh", "abi", "deh", "mo", "mo bai"},
		TLDs:        []string{".gy"},
		Currencies:  []string{"gyd"},
		MusicGenres: []string{"chutney soca", "reggae"},
		Weight:      1.1,
	},
	{
		Name: "Haiti",
		PhoneCodes: []string{"+509", "509-", "509 "},
		Cities:      []string{"port-au-prince", "cap-haïtien", "cap haitien", "gonaïves", "gonaives", "jacmel", "les cayes"},
		Keywords:    []string{"haiti", "haitian", "ayiti", "509", "kreyòl"},
		Hashtags:    []string{"haiti", "haitian", "ayiti", "509", "haitianpride"},
		Slang:       []string{"mèsi", "bonjou", "bonswa", "kijan ou rele", "pa konprann"},
		TLDs:        []string{".ht"},
		Currencies:  []string{"htg", "gourde"},
		MusicGenres: []string{"kompa", "rara", "zouk"},
		Weight:      1.1,
	},
	{
		Name: "United Kingdom",
		PhoneCodes: []string{"+44", "+44 7", "+447"},
		Cities: []string{
			"london", "manchester", "birmingham", "liverpool",
			"leeds", "glasgow", "edinburgh", "bristol", "sheffield",
			"nottingham", "leicester", "coventry", "bradford", "newcastle",
		},
		Keywords: []string{"uk", "britain", "british", "england", "scotland", "wales", "northern ireland"},
		Hashtags: []string{"uk", "london", "britain", "uklife", "britishproblems", "uktiktok"},
		Slang:    []string{"innit", "bruv", "mate", "bloke", "cheers", "init", "bare", "mandem", "roadman", "peng"},
		TLDs:     []string{".uk", ".co.uk"},
		Currencies: []string{"gbp", "£"},
		MusicGenres: []string{"grime", "uk drill", "garage"},
		Weight:      1.0,
	},
	{
		Name: "United States",
		PhoneCodes: []string{"+1"},
		Cities: []string{
			"new york", "los angeles", "chicago", "houston", "miami",
			"atlanta", "bronx", "brooklyn", "queens", "manhattan",
			"philadelphia", "dallas", "phoenix", "san antonio",
			"las vegas", "orlando", "boston", "seattle",
		},
		Keywords:    []string{"usa", "america", "american"},
		Hashtags:    []string{"usa", "america", "nyc", "la", "atl", "miami", "chicago", "houston"},
		TLDs:        []string{".us"},
		Currencies:  []string{"usd"},
		MusicGenres: []string{"hip hop", "rap", "r&b", "country"},
		Weight:      0.8,
	},
	{
		Name: "Canada",
		PhoneCodes: []string{"+1"},
		Cities: []string{
			"toronto", "vancouver", "montreal", "calgary", "ottawa",
			"edmonton", "winnipeg", "hamilton", "brampton", "surrey", "mississauga",
		},
		Keywords:    []string{"canada", "canadian", "toronto", "vancouver"},
		Hashtags:    []string{"canada", "canadian", "toronto", "vancouver", "calgary"},
		TLDs:        []string{".ca"},
		Currencies:  []string{"cad"},
		MusicGenres: []string{},
		Weight:      0.8,
	},
	{
		Name: "Nigeria",
		PhoneCodes: []string{"+234", "0803", "0806", "0802", "0814", "0816", "0901"},
		Cities: []string{
			"lagos", "abuja", "kano", "ibadan", "port harcourt",
			"benin city", "kaduna", "maiduguri", "zaria", "aba", "enugu",
		},
		Keywords:    []string{"nigeria", "nigerian", "naija", "9ja", "nollywood"},
		Hashtags:    []string{"nigeria", "nigerian", "naija", "9ja", "afrobeats", "nollywood"},
		Slang:       []string{"abeg", "oga", "no be", "na so", "wahala", "abi", "nna", "werey", "sharp sharp"},
		TLDs:        []string{".ng"},
		Currencies:  []string{"ngn", "naira", "₦"},
		MusicGenres: []string{"afrobeats", "afrobeat", "highlife", "fuji"},
		Weight:      1.0,
	},
	{
		Name: "Ghana",
		PhoneCodes: []string{"+233", "0244", "0542", "0500"},
		Cities:      []string{"accra", "kumasi", "tamale", "cape coast", "tema", "sekondi"},
		Keywords:    []string{"ghana", "ghanaian", "gh", "gold coast"},
		Hashtags:    []string{"ghana", "ghanaian", "accra", "kumasi", "afrobeats", "ghanagram"},
		Slang:       []string{"bre", "yawa", "chale", "ei", "tweaa", "wo hu"},
		TLDs:        []string{".gh"},
		Currencies:  []string{"ghs", "cedi"},
		MusicGenres: []string{"highlife", "afrobeats", "hiplife"},
		Weight:      1.0,
	},
}

// ContainsAny checks if text contains any of the patterns (case-insensitive).
// Returns true and the matched pattern.
func ContainsAny(text string, patterns []string) (bool, string) {
	lower := strings.ToLower(text)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true, p
		}
	}
	return false, ""
}
