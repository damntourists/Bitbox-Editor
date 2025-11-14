package font

// FontFamily represents a font family
type FontFamily string

const (
	FontFamilySatoshi  FontFamily = "Satoshi"
	FontFamilyFiraCode FontFamily = "FiraCode"
	FontFamilyLucide   FontFamily = "Lucide"
	FontFamilyHegarty  FontFamily = "Hegarty"
	FontFamilyBebas    FontFamily = "Bebas"
	FontFamilyPacifico FontFamily = "Pacifico"
	FontFamilyNabla    FontFamily = "Nabla"
)

// FontWeight represents a font weight
type FontWeight string

const (
	FontWeightBlack    FontWeight = "Black"
	FontWeightBold     FontWeight = "Bold"
	FontWeightSemiBold FontWeight = "SemiBold"
	FontWeightMedium   FontWeight = "Medium"
	FontWeightRegular  FontWeight = "Regular"
	FontWeightRetina   FontWeight = "Retina"
	FontWeightLight    FontWeight = "Light"
	FontWeightItalic   FontWeight = "Italic"
)

// fontPaths maps font families and weights to resource paths
var fontPaths = map[FontFamily]map[FontWeight]string{
	FontFamilySatoshi: {
		FontWeightBlack:   "fonts/Satoshi_Complete/Satoshi-Black.ttf",
		FontWeightBold:    "fonts/Satoshi_Complete/Satoshi-Bold.ttf",
		FontWeightMedium:  "fonts/Satoshi_Complete/Satoshi-Medium.ttf",
		FontWeightRegular: "fonts/Satoshi_Complete/Satoshi-Regular.ttf",
		FontWeightLight:   "fonts/Satoshi_Complete/Satoshi-Light.ttf",
		FontWeightItalic:  "fonts/Satoshi_Complete/Satoshi-Italic.ttf",
	},
	FontFamilyFiraCode: {
		FontWeightBold:     "fonts/FiraCode/FiraCode-Bold.ttf",
		FontWeightSemiBold: "fonts/FiraCode/FiraCode-SemiBold.ttf",
		FontWeightMedium:   "fonts/FiraCode/FiraCode-Medium.ttf",
		FontWeightRegular:  "fonts/FiraCode/FiraCode-Regular.ttf",
		FontWeightRetina:   "fonts/FiraCode/FiraCode-Retina.ttf",
		FontWeightLight:    "fonts/FiraCode/FiraCode-Light.ttf",
	},
	FontFamilyLucide: {
		FontWeightRegular: "fonts/Lucide/lucide.ttf",
	},
	FontFamilyBebas: {
		FontWeightRegular: "fonts/Bebas/Bebas-Regular.ttf",
	},
	FontFamilyHegarty: {
		FontWeightRegular: "fonts/Hegarty/Hegarty-Regular.ttf",
	},
	FontFamilyPacifico: {
		FontWeightRegular: "fonts/Pacifico/Pacifico-Regular.ttf",
	},
	FontFamilyNabla: {
		FontWeightRegular: "fonts/Nabla/Nabla-Regular-VariableFont_EDPT,EHLT.ttf",
	},
}
