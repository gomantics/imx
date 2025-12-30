package xmp

var wellKnownPrefixes = map[string]string{
	"adobe:ns:meta/":                                       "x", // XMP metadata namespace
	"http://ns.adobe.com/xap/1.0/":                         "xmp",
	"http://ns.adobe.com/xap/1.0/mm/":                      "xmpMM",
	"http://ns.adobe.com/xap/1.0/st/":                      "xmpST",
	"http://ns.adobe.com/xap/1.0/rights/":                  "xmpRights",
	"http://purl.org/dc/elements/1.1/":                     "dc",
	"http://iptc.org/std/Iptc4xmpCore/1.0/xmlns/":          "Iptc4xmpCore",
	"http://ns.adobe.com/photoshop/1.0/":                   "photoshop",
	"http://ns.adobe.com/tiff/1.0/":                        "tiff",
	"http://ns.adobe.com/exif/1.0/":                        "exif",
	"http://ns.adobe.com/camera-raw-settings/1.0/":         "crs",
	"http://www.metadataworkinggroup.com/schemas/regions/": "mwg-rs",
	"http://ns.apple.com/faceinfo/1.0/":                    "apple-fi",
	"http://ns.adobe.com/xmp/sType/Area#":                  "stArea",
	"http://ns.adobe.com/xap/1.0/sType/Dimensions#":        "stDim",
	"http://ns.adobe.com/xap/1.0/sType/ResourceEvent#":     "stEvt",
}
