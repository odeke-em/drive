//line cc.y:34
package cc

import __yyfmt__ "fmt"

//line cc.y:34
type typeClass struct {
	c Storage
	q TypeQual
	t *Type
}

type idecor struct {
	d func(*Type) (*Type, string)
	i *Init
}

//line cc.y:49
type yySymType struct {
	yys      int
	abdecor  func(*Type) *Type
	decl     *Decl
	decls    []*Decl
	decor    func(*Type) (*Type, string)
	decors   []func(*Type) (*Type, string)
	expr     *Expr
	exprs    []*Expr
	idec     idecor
	idecs    []idecor
	init     *Init
	inits    []*Init
	label    *Label
	labels   []*Label
	span     Span
	prefix   *Prefix
	prefixes []*Prefix
	stmt     *Stmt
	stmts    []*Stmt
	str      string
	strs     []string
	tc       typeClass
	tk       TypeKind
	typ      *Type
}

const tokARGBEGIN = 57346
const tokARGEND = 57347
const tokAUTOLIB = 57348
const tokSET = 57349
const tokUSED = 57350
const tokAuto = 57351
const tokBreak = 57352
const tokCase = 57353
const tokChar = 57354
const tokConst = 57355
const tokContinue = 57356
const tokDefault = 57357
const tokDo = 57358
const tokDotDotDot = 57359
const tokDouble = 57360
const tokEnum = 57361
const tokError = 57362
const tokExtern = 57363
const tokFloat = 57364
const tokFor = 57365
const tokGoto = 57366
const tokIf = 57367
const tokInline = 57368
const tokInt = 57369
const tokLitChar = 57370
const tokLong = 57371
const tokName = 57372
const tokNumber = 57373
const tokOffsetof = 57374
const tokRegister = 57375
const tokReturn = 57376
const tokShort = 57377
const tokSigned = 57378
const tokStatic = 57379
const tokStruct = 57380
const tokSwitch = 57381
const tokTypeName = 57382
const tokTypedef = 57383
const tokUnion = 57384
const tokUnsigned = 57385
const tokVaArg = 57386
const tokVoid = 57387
const tokVolatile = 57388
const tokWhile = 57389
const tokString = 57390
const tokShift = 57391
const tokElse = 57392
const tokAddEq = 57393
const tokSubEq = 57394
const tokMulEq = 57395
const tokDivEq = 57396
const tokModEq = 57397
const tokLshEq = 57398
const tokRshEq = 57399
const tokAndEq = 57400
const tokXorEq = 57401
const tokOrEq = 57402
const tokOrOr = 57403
const tokAndAnd = 57404
const tokEqEq = 57405
const tokNotEq = 57406
const tokLtEq = 57407
const tokGtEq = 57408
const tokLsh = 57409
const tokRsh = 57410
const tokCast = 57411
const tokSizeof = 57412
const tokUnary = 57413
const tokDec = 57414
const tokInc = 57415
const tokArrow = 57416
const startExpr = 57417
const startProg = 57418
const tokEOF = 57419

var yyToknames = []string{
	"tokARGBEGIN",
	"tokARGEND",
	"tokAUTOLIB",
	"tokSET",
	"tokUSED",
	"tokAuto",
	"tokBreak",
	"tokCase",
	"tokChar",
	"tokConst",
	"tokContinue",
	"tokDefault",
	"tokDo",
	"tokDotDotDot",
	"tokDouble",
	"tokEnum",
	"tokError",
	"tokExtern",
	"tokFloat",
	"tokFor",
	"tokGoto",
	"tokIf",
	"tokInline",
	"tokInt",
	"tokLitChar",
	"tokLong",
	"tokName",
	"tokNumber",
	"tokOffsetof",
	"tokRegister",
	"tokReturn",
	"tokShort",
	"tokSigned",
	"tokStatic",
	"tokStruct",
	"tokSwitch",
	"tokTypeName",
	"tokTypedef",
	"tokUnion",
	"tokUnsigned",
	"tokVaArg",
	"tokVoid",
	"tokVolatile",
	"tokWhile",
	"tokString",
	"tokShift",
	"tokElse",
	" {",
	" ,",
	" =",
	"tokAddEq",
	"tokSubEq",
	"tokMulEq",
	"tokDivEq",
	"tokModEq",
	"tokLshEq",
	"tokRshEq",
	"tokAndEq",
	"tokXorEq",
	"tokOrEq",
	" ?",
	" :",
	"tokOrOr",
	"tokAndAnd",
	" |",
	" ^",
	" &",
	"tokEqEq",
	"tokNotEq",
	" <",
	" >",
	"tokLtEq",
	"tokGtEq",
	"tokLsh",
	"tokRsh",
	" +",
	" -",
	" *",
	" /",
	" %",
	"tokCast",
	" !",
	" ~",
	"tokSizeof",
	"tokUnary",
	" .",
	" [",
	" ]",
	" (",
	" )",
	"tokDec",
	"tokInc",
	"tokArrow",
	"startExpr",
	"startProg",
	"tokEOF",
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 118,
	52, 100,
	101, 100,
	-2, 180,
	-1, 136,
	51, 171,
	-2, 145,
	-1, 138,
	51, 171,
	-2, 150,
	-1, 238,
	101, 206,
	-2, 170,
	-1, 269,
	65, 171,
	-2, 91,
}

const yyNprod = 216
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1452

var yyAct = []int{

	309, 7, 112, 120, 342, 302, 259, 31, 225, 266,
	213, 279, 111, 240, 98, 99, 100, 101, 102, 103,
	104, 105, 106, 193, 196, 219, 117, 343, 237, 49,
	109, 5, 223, 123, 4, 217, 308, 133, 131, 377,
	129, 136, 138, 375, 227, 368, 367, 32, 363, 356,
	110, 354, 337, 336, 334, 289, 118, 288, 187, 305,
	34, 296, 140, 141, 142, 143, 144, 145, 146, 147,
	148, 149, 150, 151, 152, 153, 154, 155, 156, 157,
	158, 130, 160, 161, 162, 163, 164, 165, 166, 167,
	168, 169, 170, 127, 292, 135, 244, 59, 210, 174,
	175, 249, 61, 62, 63, 64, 65, 159, 35, 295,
	3, 2, 96, 92, 380, 91, 184, 94, 93, 95,
	173, 124, 180, 96, 92, 374, 91, 124, 94, 93,
	95, 125, 190, 366, 365, 110, 364, 125, 176, 177,
	361, 128, 360, 134, 66, 67, 61, 62, 63, 64,
	65, 195, 189, 284, 188, 235, 96, 92, 189, 91,
	188, 94, 93, 95, 283, 252, 198, 197, 181, 189,
	199, 188, 121, 215, 130, 205, 203, 179, 256, 183,
	207, 178, 362, 122, 345, 344, 341, 339, 135, 257,
	333, 224, 226, 135, 230, 332, 212, 299, 115, 114,
	108, 221, 282, 6, 242, 211, 258, 207, 243, 204,
	210, 195, 224, 238, 348, 347, 291, 232, 233, 31,
	202, 216, 247, 208, 234, 221, 231, 274, 229, 290,
	376, 271, 254, 280, 281, 253, 134, 206, 60, 128,
	192, 134, 269, 246, 260, 250, 255, 226, 248, 238,
	208, 201, 200, 186, 277, 232, 116, 261, 97, 63,
	64, 65, 352, 267, 241, 263, 221, 96, 92, 57,
	91, 124, 94, 93, 95, 268, 294, 33, 185, 285,
	286, 125, 297, 301, 300, 293, 195, 270, 228, 287,
	298, 1, 52, 304, 269, 172, 57, 247, 37, 226,
	303, 11, 58, 194, 113, 230, 306, 132, 48, 56,
	126, 233, 312, 137, 139, 267, 55, 278, 317, 311,
	53, 313, 245, 338, 54, 335, 276, 119, 340, 58,
	171, 346, 272, 273, 264, 265, 239, 236, 230, 318,
	28, 26, 218, 191, 353, 29, 182, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	349, 350, 0, 0, 0, 371, 372, 373, 370, 355,
	0, 0, 357, 358, 0, 0, 0, 379, 0, 0,
	378, 381, 0, 0, 0, 0, 0, 0, 0, 0,
	369, 80, 81, 82, 83, 84, 85, 86, 87, 88,
	89, 90, 79, 359, 78, 77, 76, 75, 74, 72,
	73, 68, 69, 70, 71, 66, 67, 61, 62, 63,
	64, 65, 0, 0, 0, 0, 0, 96, 92, 0,
	91, 0, 94, 93, 95, 319, 0, 0, 316, 315,
	0, 320, 329, 0, 0, 321, 330, 322, 0, 0,
	0, 0, 0, 0, 323, 324, 325, 0, 0, 10,
	0, 331, 9, 21, 0, 326, 0, 0, 0, 0,
	327, 0, 0, 0, 0, 23, 0, 0, 328, 24,
	0, 0, 260, 74, 72, 73, 68, 69, 70, 71,
	66, 67, 61, 62, 63, 64, 65, 0, 0, 0,
	0, 13, 96, 92, 0, 91, 0, 94, 93, 95,
	14, 15, 12, 0, 0, 0, 16, 17, 20, 0,
	0, 0, 0, 22, 0, 19, 18, 0, 0, 0,
	0, 0, 314, 80, 81, 82, 83, 84, 85, 86,
	87, 88, 89, 90, 79, 0, 78, 77, 76, 75,
	74, 72, 73, 68, 69, 70, 71, 66, 67, 61,
	62, 63, 64, 65, 0, 0, 0, 0, 0, 96,
	92, 307, 91, 0, 94, 93, 95, 80, 81, 82,
	83, 84, 85, 86, 87, 88, 89, 90, 79, 0,
	78, 77, 76, 75, 74, 72, 73, 68, 69, 70,
	71, 66, 67, 61, 62, 63, 64, 65, 0, 0,
	0, 0, 0, 96, 92, 0, 91, 275, 94, 93,
	95, 214, 80, 81, 82, 83, 84, 85, 86, 87,
	88, 89, 90, 79, 0, 78, 77, 76, 75, 74,
	72, 73, 68, 69, 70, 71, 66, 67, 61, 62,
	63, 64, 65, 0, 0, 0, 0, 0, 96, 92,
	0, 91, 0, 94, 93, 95, 80, 81, 82, 83,
	84, 85, 86, 87, 88, 89, 90, 79, 0, 78,
	77, 76, 75, 74, 72, 73, 68, 69, 70, 71,
	66, 67, 61, 62, 63, 64, 65, 0, 0, 0,
	0, 0, 96, 92, 0, 91, 0, 94, 93, 95,
	52, 0, 0, 39, 57, 0, 0, 0, 0, 46,
	38, 0, 113, 45, 0, 0, 0, 56, 41, 10,
	42, 8, 9, 21, 55, 0, 40, 43, 53, 50,
	0, 36, 54, 51, 44, 23, 47, 58, 0, 24,
	77, 76, 75, 74, 72, 73, 68, 69, 70, 71,
	66, 67, 61, 62, 63, 64, 65, 0, 0, 0,
	0, 13, 96, 92, 0, 91, 0, 94, 93, 95,
	14, 15, 12, 0, 0, 0, 16, 17, 20, 0,
	0, 27, 0, 22, 52, 19, 18, 39, 57, 0,
	0, 0, 0, 46, 38, 0, 30, 45, 0, 0,
	0, 56, 41, 0, 42, 0, 0, 0, 55, 0,
	40, 43, 53, 50, 0, 36, 54, 51, 44, 52,
	47, 58, 39, 57, 0, 0, 0, 0, 46, 38,
	0, 113, 45, 0, 0, 0, 56, 41, 0, 42,
	0, 0, 0, 55, 0, 40, 43, 53, 50, 0,
	36, 54, 51, 44, 52, 47, 58, 39, 57, 0,
	0, 0, 0, 46, 38, 0, 113, 45, 0, 0,
	0, 56, 41, 0, 42, 251, 0, 0, 55, 0,
	40, 43, 53, 50, 0, 36, 54, 51, 44, 0,
	47, 58, 27, 0, 0, 52, 0, 0, 39, 57,
	0, 0, 0, 0, 46, 38, 0, 30, 45, 0,
	310, 0, 56, 41, 0, 42, 0, 0, 0, 55,
	0, 40, 43, 53, 50, 0, 36, 54, 51, 44,
	0, 47, 58, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 79, 262, 78, 77, 76, 75,
	74, 72, 73, 68, 69, 70, 71, 66, 67, 61,
	62, 63, 64, 65, 0, 0, 0, 0, 0, 96,
	92, 0, 91, 0, 94, 93, 95, 0, 0, 0,
	0, 0, 0, 0, 0, 25, 76, 75, 74, 72,
	73, 68, 69, 70, 71, 66, 67, 61, 62, 63,
	64, 65, 0, 0, 0, 0, 0, 96, 92, 0,
	91, 0, 94, 93, 95, 75, 74, 72, 73, 68,
	69, 70, 71, 66, 67, 61, 62, 63, 64, 65,
	0, 0, 0, 0, 0, 96, 92, 0, 91, 0,
	94, 93, 95, 72, 73, 68, 69, 70, 71, 66,
	67, 61, 62, 63, 64, 65, 10, 0, 8, 9,
	21, 96, 92, 0, 91, 0, 94, 93, 95, 0,
	0, 0, 23, 0, 0, 0, 24, 0, 0, 209,
	0, 0, 73, 68, 69, 70, 71, 66, 67, 61,
	62, 63, 64, 65, 0, 0, 0, 0, 13, 96,
	92, 0, 91, 0, 94, 93, 95, 14, 15, 12,
	0, 0, 0, 16, 17, 20, 0, 280, 281, 0,
	22, 0, 19, 18, 68, 69, 70, 71, 66, 67,
	61, 62, 63, 64, 65, 10, 0, 8, 9, 21,
	96, 92, 0, 91, 0, 94, 93, 95, 0, 0,
	0, 23, 0, 0, 0, 24, 0, 0, 209, 0,
	0, 0, 10, 0, 8, 9, 21, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 13, 23, 0,
	0, 0, 24, 0, 0, 0, 14, 15, 12, 0,
	0, 0, 16, 17, 20, 0, 0, 0, 0, 22,
	0, 19, 18, 0, 13, 0, 0, 10, 0, 8,
	9, 21, 0, 14, 15, 12, 0, 0, 0, 16,
	17, 20, 0, 23, 0, 0, 22, 24, 19, 18,
	10, 0, 8, 9, 21, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 23, 0, 0, 13,
	24, 0, 0, 209, 0, 0, 0, 0, 14, 15,
	12, 0, 0, 0, 16, 17, 20, 0, 0, 0,
	0, 107, 0, 19, 18, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 16, 17, 20,
	0, 0, 0, 0, 22, 52, 19, 18, 39, 57,
	0, 0, 0, 222, 46, 38, 0, 113, 45, 0,
	0, 0, 56, 41, 0, 42, 220, 0, 0, 55,
	0, 40, 43, 53, 50, 0, 36, 54, 51, 44,
	351, 47, 58, 0, 52, 0, 0, 39, 57, 0,
	0, 0, 0, 46, 38, 0, 113, 45, 0, 0,
	0, 56, 41, 0, 42, 0, 0, 0, 55, 0,
	40, 43, 53, 50, 0, 36, 54, 51, 44, 52,
	47, 58, 39, 57, 0, 0, 0, 0, 46, 38,
	0, 113, 45, 0, 0, 0, 56, 41, 0, 42,
	0, 0, 0, 55, 0, 40, 43, 53, 50, 0,
	36, 54, 51, 44, 52, 47, 58, 39, 57, 0,
	0, 0, 0, 46, 0, 0, 113, 45, 0, 0,
	0, 56, 41, 0, 42, 0, 0, 0, 55, 0,
	40, 43, 53, 0, 0, 0, 54, 0, 44, 0,
	47, 58,
}
var yyPact = []int{

	13, -1000, -1000, 1144, 896, -2, 186, 613, -1000, -1000,
	-1000, 210, 1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144,
	1189, 108, 701, 107, -1000, -1000, -1000, 106, -1000, -1000,
	208, 91, 1370, 283, 1405, -1000, -1000, 241, 241, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144,
	1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144,
	1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144, 1144,
	1144, 1144, 1144, -1000, -1000, 241, 241, -1000, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 701, 1370, 88,
	84, 87, -1000, -1000, 1144, 248, 202, -43, 79, 188,
	-1000, 256, 91, -1000, -1000, -1000, 283, 1405, -1000, -1000,
	283, -1000, 1405, -1000, -1000, -1000, -1000, 201, -1000, 200,
	613, 178, 178, 34, 34, 34, 23, 23, 67, 67,
	67, 67, 1020, 1061, 982, 413, 956, 928, 683, 155,
	613, 613, 613, 613, 613, 613, 613, 613, 613, 613,
	613, 83, 186, 118, -1000, -1000, 82, 185, 1117, -1000,
	120, 256, 104, 87, 569, 80, -1000, -1000, 1296, 1144,
	1117, 1370, 91, 91, 256, -1000, 62, -1000, -1000, -1000,
	1370, 234, 1144, -1000, -1000, 1212, 1144, 34, -1000, -4,
	1144, 87, 1296, 8, 1370, -1000, 785, 72, 183, -1000,
	-1000, 97, -1000, 115, 613, -1000, 613, -1000, 193, -1000,
	91, -1000, 79, 68, -1000, -1000, 855, -1000, 91, 179,
	-1000, 174, 890, 524, -1000, 1038, 111, 120, 71, -1000,
	60, -1000, -1000, 1296, 120, 68, 256, 97, -1000, -1000,
	-1000, -44, -1000, -1000, -46, 177, -1000, 68, 151, -1000,
	-6, 234, -1000, -1000, 1144, -1000, 9, -1000, 144, -1000,
	241, 1144, -1000, -1000, -1000, -1000, 97, -1000, -1000, -1000,
	91, 1144, -1000, -1000, 613, -1000, -41, 1117, -1000, -1000,
	-1000, 480, 820, -1000, 613, -1000, -1000, -1000, -1000, -1000,
	-1000, 431, -1000, -1000, -1000, 103, 98, -1000, -47, -1000,
	-48, -49, -1000, 95, 241, 94, 1144, 93, 92, 1144,
	150, 149, 1144, 1144, -1000, 1335, -1000, -1000, 215, 1144,
	-50, 1144, -52, -1000, 1144, 1144, 338, -1000, -1000, 49,
	47, -1000, 90, -53, -1000, 43, -1000, 41, 40, -1000,
	-55, -56, 1144, 1144, -1000, -1000, -1000, -1000, -1000, 32,
	-58, 180, -1000, -1000, -62, 1144, -1000, -1000, 21, -1000,
	-1000, -1000,
}
var yyPgo = []int{

	0, 10, 346, 25, 345, 13, 36, 343, 342, 35,
	34, 341, 340, 28, 337, 336, 24, 9, 335, 334,
	1, 32, 27, 4, 333, 332, 203, 330, 33, 327,
	26, 8, 326, 44, 322, 321, 319, 11, 317, 312,
	6, 0, 5, 308, 29, 60, 108, 37, 3, 275,
	47, 40, 307, 38, 303, 23, 301, 2, 298, 30,
	12, 277, 291, 289, 288, 287, 282,
}
var yyR1 = []int{

	0, 62, 62, 10, 10, 10, 22, 20, 20, 20,
	20, 20, 20, 20, 20, 20, 20, 20, 20, 20,
	20, 20, 20, 20, 20, 20, 20, 20, 20, 20,
	20, 20, 20, 20, 20, 20, 20, 20, 20, 20,
	20, 20, 20, 20, 20, 20, 20, 20, 20, 20,
	20, 20, 20, 20, 20, 20, 20, 20, 20, 20,
	42, 42, 42, 63, 40, 35, 35, 35, 41, 39,
	39, 39, 39, 39, 39, 39, 39, 39, 39, 39,
	39, 39, 39, 39, 39, 1, 1, 1, 2, 2,
	2, 16, 16, 16, 16, 16, 3, 3, 3, 3,
	28, 28, 43, 43, 43, 43, 43, 43, 44, 44,
	45, 45, 45, 45, 45, 45, 45, 45, 45, 46,
	46, 47, 47, 61, 57, 57, 57, 57, 57, 60,
	59, 6, 12, 11, 11, 11, 64, 4, 48, 48,
	58, 58, 17, 17, 13, 61, 61, 37, 20, 20,
	61, 61, 5, 24, 31, 31, 33, 33, 33, 34,
	34, 32, 32, 37, 66, 66, 65, 65, 38, 38,
	49, 49, 23, 23, 21, 21, 26, 26, 27, 27,
	7, 7, 36, 36, 8, 8, 9, 9, 29, 29,
	30, 30, 54, 54, 55, 55, 50, 50, 51, 51,
	52, 52, 53, 53, 18, 18, 19, 19, 14, 14,
	25, 25, 15, 15, 56, 56,
}
var yyR2 = []int{

	0, 3, 3, 0, 2, 5, 1, 1, 1, 1,
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 5,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	4, 6, 4, 4, 3, 4, 4, 2, 2, 6,
	0, 2, 2, 0, 4, 3, 2, 2, 2, 1,
	5, 5, 1, 2, 3, 2, 2, 7, 9, 3,
	5, 7, 3, 5, 5, 0, 3, 1, 4, 4,
	3, 1, 3, 3, 4, 4, 1, 2, 2, 1,
	1, 3, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 3, 3, 2, 2, 1,
	2, 3, 3, 1, 1, 5, 0, 5, 1, 1,
	1, 1, 1, 3, 3, 2, 5, 2, 3, 3,
	2, 6, 2, 2, 1, 1, 2, 4, 5, 0,
	3, 1, 3, 3, 0, 1, 0, 1, 1, 2,
	0, 1, 0, 1, 0, 1, 1, 3, 0, 1,
	0, 2, 0, 2, 1, 3, 0, 1, 1, 3,
	0, 1, 1, 2, 0, 1, 1, 2, 0, 1,
	1, 2, 0, 1, 1, 3, 0, 1, 1, 2,
	0, 1, 1, 3, 1, 2,
}
var yyChk = []int{

	-1000, -62, 98, 97, -10, -22, -26, -20, 30, 31,
	28, -56, 81, 70, 79, 80, 85, 86, 95, 94,
	87, 32, 92, 44, 48, 99, -11, 6, -12, -4,
	21, -57, -50, -61, -45, -46, 40, -58, 19, 12,
	35, 27, 29, 36, 43, 22, 18, 45, -43, -44,
	38, 42, 9, 37, 41, 33, 26, 13, 46, 99,
	52, 79, 80, 81, 82, 83, 77, 78, 73, 74,
	75, 76, 71, 72, 70, 69, 68, 67, 66, 64,
	53, 54, 55, 56, 57, 58, 59, 60, 61, 62,
	63, 92, 90, 95, 94, 96, 89, 48, -20, -20,
	-20, -20, -20, -20, -20, -20, -20, 92, 92, -59,
	-22, -60, -57, 21, 92, 92, 48, -30, -16, -29,
	-48, 81, 92, -28, 30, 40, -61, -45, -46, -51,
	-50, -53, -52, -47, -46, -45, -48, -49, -48, -49,
	-20, -20, -20, -20, -20, -20, -20, -20, -20, -20,
	-20, -20, -20, -20, -20, -20, -20, -20, -20, -22,
	-20, -20, -20, -20, -20, -20, -20, -20, -20, -20,
	-20, -27, -26, -22, -48, -48, -59, -59, 93, 93,
	-1, 81, -2, 92, -20, 30, 51, 101, 92, 90,
	53, -7, 52, -55, -54, -44, -16, -51, -53, -47,
	51, 51, 65, 93, 91, 93, 52, -20, -33, 51,
	90, -55, 92, -1, 52, 93, -10, -9, -8, -3,
	30, -60, 17, -21, -20, -31, -20, -33, -64, -6,
	-57, -28, -16, -16, -44, 93, -14, -13, -60, -15,
	-5, 30, -20, -20, 100, -34, -21, -1, -9, 93,
	-59, 100, 93, 52, -1, -16, 81, 92, 91, -40,
	51, -30, 100, -13, -19, -18, -17, -16, -49, -48,
	-65, 52, -25, -24, 53, 93, -32, -31, -38, -37,
	89, 90, 91, 93, 93, -3, -55, -63, 101, 101,
	52, 65, 100, -5, -20, 100, 52, -66, -37, 53,
	-48, -20, -42, -17, -20, 100, -31, 91, -6, -41,
	100, -36, -39, -35, 101, 8, 7, -40, -22, 4,
	10, 14, 16, 23, 24, 25, 34, 39, 47, 11,
	15, 30, 92, 92, 101, -42, 101, 101, -41, 92,
	-48, 92, -23, -22, 92, 92, -20, 65, 65, -22,
	-22, 5, 47, -23, 101, -22, 101, -22, -22, 65,
	93, 93, 92, 101, 93, 93, 93, 101, 101, -22,
	-23, -41, -41, -41, 93, 101, 50, 101, -23, -41,
	93, -41,
}
var yyDef = []int{

	0, -2, 3, 0, 0, 0, 6, 176, 7, 8,
	9, 10, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 214, 1, 4, 0, 133, 134,
	104, 190, 124, 198, 202, 196, 123, 170, 170, 110,
	111, 112, 113, 114, 115, 116, 117, 118, 119, 120,
	140, 141, 102, 103, 105, 106, 107, 108, 109, 2,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 178, 0, 57, 58, 0, 0, 215, 41, 42,
	43, 44, 45, 46, 47, 48, 49, 0, 0, 0,
	0, 85, 129, 104, 0, 0, 0, 0, -2, 191,
	91, 194, 0, 188, 138, 139, 198, 202, 197, 127,
	199, 128, 203, 200, 121, 122, -2, 0, -2, 0,
	177, 11, 12, 13, 14, 15, 16, 17, 18, 19,
	20, 21, 22, 23, 24, 25, 26, 27, 28, 0,
	30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
	40, 0, 179, 0, 148, 149, 0, 0, 0, 54,
	130, 194, 87, 85, 0, 0, 3, 132, 186, 174,
	0, 136, 0, 0, 195, 192, 0, 125, 126, 201,
	0, 0, 0, 55, 56, 50, 0, 52, 53, 159,
	174, 85, 186, 0, 0, 5, 0, 0, 187, 184,
	96, 85, 99, 0, 175, 101, 154, 155, 0, 181,
	190, 189, 100, 92, 193, 93, 0, 208, -2, 166,
	212, 210, 29, 0, 156, 0, 0, 86, 0, 90,
	0, 135, 94, 0, 97, 98, 194, 85, 95, 137,
	63, 0, 146, 209, 0, 207, 204, 142, 0, -2,
	0, 167, 152, 211, 0, 51, 0, 161, 164, 168,
	0, 0, 89, 88, 59, 185, 85, 60, 131, 144,
	170, 0, 151, 213, 153, 157, 160, 0, 169, 165,
	147, 0, 182, 205, 143, 158, 162, 163, 61, 62,
	64, 0, 68, 183, 69, 0, 0, 72, 0, 60,
	0, 0, 182, 0, 0, 0, 172, 0, 0, 0,
	0, 7, 0, 0, 73, 182, 75, 76, 0, 172,
	0, 0, 0, 173, 0, 0, 0, 66, 67, 0,
	0, 74, 0, 0, 79, 0, 82, 0, 0, 65,
	0, 0, 0, 172, 182, 182, 182, 70, 71, 0,
	0, 80, 83, 84, 0, 172, 182, 77, 0, 81,
	182, 78,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 85, 3, 3, 3, 83, 70, 3,
	92, 93, 81, 79, 52, 80, 89, 82, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 65, 101,
	73, 53, 74, 64, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 90, 3, 91, 69, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 51, 68, 100, 86,
}
var yyTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 49, 50, 54,
	55, 56, 57, 58, 59, 60, 61, 62, 63, 66,
	67, 71, 72, 75, 76, 77, 78, 84, 87, 88,
	94, 95, 96, 97, 98, 99,
}
var yyTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

const yyFlag = -1000

func yyTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(yyToknames) {
		if yyToknames[c-4] != "" {
			return yyToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yylex1(lex yyLexer, lval *yySymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		c = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			c = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		c = yyTok3[i+0]
		if c == char {
			c = yyTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(c), uint(char))
	}
	return c
}

func yyParse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yychar), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar = yylex1(yylex, &yylval)
	}
	yyn += yychar
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yychar { /* valid shift */
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yychar {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error("syntax error")
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yychar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}
			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		//line cc.y:185
		{
			yylex.(*lexer).prog = &Prog{Decls: yyS[yypt-1].decls}
			return 0
		}
	case 2:
		//line cc.y:190
		{
			yylex.(*lexer).expr = yyS[yypt-1].expr
			return 0
		}
	case 3:
		//line cc.y:196
		{
			yyVAL.span = Span{}
			yyVAL.decls = nil
		}
	case 4:
		//line cc.y:201
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.decls = append(yyS[yypt-1].decls, yyS[yypt-0].decls...)
		}
	case 5:
		//line cc.y:206
		{
		}
	case 6:
		//line cc.y:211
		{
			yyVAL.span = yyS[yypt-0].span
			if len(yyS[yypt-0].exprs) == 1 {
				yyVAL.expr = yyS[yypt-0].exprs[0]
				break
			}
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Comma, List: yyS[yypt-0].exprs}
		}
	case 7:
		//line cc.y:222
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Name, Text: yyS[yypt-0].str, XDecl: yyS[yypt-0].decl}
		}
	case 8:
		//line cc.y:227
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Number, Text: yyS[yypt-0].str}
		}
	case 9:
		//line cc.y:232
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Number, Text: yyS[yypt-0].str}
		}
	case 10:
		//line cc.y:237
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: String, Texts: yyS[yypt-0].strs}
		}
	case 11:
		//line cc.y:242
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Add, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 12:
		//line cc.y:247
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Sub, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 13:
		//line cc.y:252
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Mul, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 14:
		//line cc.y:257
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Div, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 15:
		//line cc.y:262
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Mod, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 16:
		//line cc.y:267
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Lsh, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 17:
		//line cc.y:272
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Rsh, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 18:
		//line cc.y:277
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Lt, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 19:
		//line cc.y:282
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Gt, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 20:
		//line cc.y:287
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: LtEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 21:
		//line cc.y:292
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: GtEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 22:
		//line cc.y:297
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: EqEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 23:
		//line cc.y:302
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: NotEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 24:
		//line cc.y:307
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: And, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 25:
		//line cc.y:312
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Xor, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 26:
		//line cc.y:317
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Or, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 27:
		//line cc.y:322
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: AndAnd, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 28:
		//line cc.y:327
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: OrOr, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 29:
		//line cc.y:332
		{
			yyVAL.span = span(yyS[yypt-4].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Cond, List: []*Expr{yyS[yypt-4].expr, yyS[yypt-2].expr, yyS[yypt-0].expr}}
		}
	case 30:
		//line cc.y:337
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Eq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 31:
		//line cc.y:342
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: AddEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 32:
		//line cc.y:347
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: SubEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 33:
		//line cc.y:352
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: MulEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 34:
		//line cc.y:357
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: DivEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 35:
		//line cc.y:362
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: ModEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 36:
		//line cc.y:367
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: LshEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 37:
		//line cc.y:372
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: RshEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 38:
		//line cc.y:377
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: AndEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 39:
		//line cc.y:382
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: XorEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 40:
		//line cc.y:387
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: OrEq, Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 41:
		//line cc.y:392
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Indir, Left: yyS[yypt-0].expr}
		}
	case 42:
		//line cc.y:397
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Addr, Left: yyS[yypt-0].expr}
		}
	case 43:
		//line cc.y:402
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Plus, Left: yyS[yypt-0].expr}
		}
	case 44:
		//line cc.y:407
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Minus, Left: yyS[yypt-0].expr}
		}
	case 45:
		//line cc.y:412
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Not, Left: yyS[yypt-0].expr}
		}
	case 46:
		//line cc.y:417
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Twid, Left: yyS[yypt-0].expr}
		}
	case 47:
		//line cc.y:422
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: PreInc, Left: yyS[yypt-0].expr}
		}
	case 48:
		//line cc.y:427
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: PreDec, Left: yyS[yypt-0].expr}
		}
	case 49:
		//line cc.y:432
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: SizeofExpr, Left: yyS[yypt-0].expr}
		}
	case 50:
		//line cc.y:437
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: SizeofType, Type: yyS[yypt-1].typ}
		}
	case 51:
		//line cc.y:442
		{
			yyVAL.span = span(yyS[yypt-5].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Offsetof, Type: yyS[yypt-3].typ, Left: yyS[yypt-1].expr}
		}
	case 52:
		//line cc.y:447
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Cast, Type: yyS[yypt-2].typ, Left: yyS[yypt-0].expr}
		}
	case 53:
		//line cc.y:452
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: CastInit, Type: yyS[yypt-2].typ, Init: &Init{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Braced: yyS[yypt-0].inits}}
		}
	case 54:
		//line cc.y:457
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Paren, Left: yyS[yypt-1].expr}
		}
	case 55:
		//line cc.y:462
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Call, Left: yyS[yypt-3].expr, List: yyS[yypt-1].exprs}
		}
	case 56:
		//line cc.y:467
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Index, Left: yyS[yypt-3].expr, Right: yyS[yypt-1].expr}
		}
	case 57:
		//line cc.y:472
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: PostInc, Left: yyS[yypt-1].expr}
		}
	case 58:
		//line cc.y:477
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: PostDec, Left: yyS[yypt-1].expr}
		}
	case 59:
		//line cc.y:482
		{
			yyVAL.span = span(yyS[yypt-5].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: VaArg, Left: yyS[yypt-3].expr, Type: yyS[yypt-1].typ}
		}
	case 60:
		//line cc.y:488
		{
			yyVAL.span = Span{}
			yyVAL.stmts = nil
		}
	case 61:
		//line cc.y:493
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.stmts = yyS[yypt-1].stmts
			for _, d := range yyS[yypt-0].decls {
				yyVAL.stmts = append(yyVAL.stmts, &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: StmtDecl, Decl: d})
			}
		}
	case 62:
		//line cc.y:501
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.stmts = append(yyS[yypt-1].stmts, yyS[yypt-0].stmt)
		}
	case 63:
		//line cc.y:508
		{
			yylex.(*lexer).pushScope()
		}
	case 64:
		//line cc.y:512
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			yylex.(*lexer).popScope()
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Block, Block: yyS[yypt-1].stmts}
		}
	case 65:
		//line cc.y:520
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.label = &Label{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Case, Expr: yyS[yypt-1].expr}
		}
	case 66:
		//line cc.y:525
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.label = &Label{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Default}
		}
	case 67:
		//line cc.y:530
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.label = &Label{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: LabelName, Name: yyS[yypt-1].str}
		}
	case 68:
		//line cc.y:537
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.stmt = yyS[yypt-0].stmt
			yyVAL.stmt.Labels = yyS[yypt-1].labels
		}
	case 69:
		//line cc.y:545
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Empty}
		}
	case 70:
		//line cc.y:550
		{
			yyVAL.span = yyS[yypt-4].span
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Empty}
		}
	case 71:
		//line cc.y:555
		{
			yyVAL.span = yyS[yypt-4].span
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Empty}
		}
	case 72:
		//line cc.y:560
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.stmt = yyS[yypt-0].stmt
		}
	case 73:
		//line cc.y:565
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: StmtExpr, Expr: yyS[yypt-1].expr}
		}
	case 74:
		//line cc.y:570
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: ARGBEGIN, Block: yyS[yypt-1].stmts}
		}
	case 75:
		//line cc.y:575
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Break}
		}
	case 76:
		//line cc.y:580
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Continue}
		}
	case 77:
		//line cc.y:585
		{
			yyVAL.span = span(yyS[yypt-6].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Do, Body: yyS[yypt-5].stmt, Expr: yyS[yypt-2].expr}
		}
	case 78:
		//line cc.y:590
		{
			yyVAL.span = span(yyS[yypt-8].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span},
				Op:   For,
				Pre:  yyS[yypt-6].expr,
				Expr: yyS[yypt-4].expr,
				Post: yyS[yypt-2].expr,
				Body: yyS[yypt-0].stmt,
			}
		}
	case 79:
		//line cc.y:601
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Goto, Text: yyS[yypt-1].str}
		}
	case 80:
		//line cc.y:606
		{
			yyVAL.span = span(yyS[yypt-4].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: If, Expr: yyS[yypt-2].expr, Body: yyS[yypt-0].stmt}
		}
	case 81:
		//line cc.y:611
		{
			yyVAL.span = span(yyS[yypt-6].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: If, Expr: yyS[yypt-4].expr, Body: yyS[yypt-2].stmt, Else: yyS[yypt-0].stmt}
		}
	case 82:
		//line cc.y:616
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Return, Expr: yyS[yypt-1].expr}
		}
	case 83:
		//line cc.y:621
		{
			yyVAL.span = span(yyS[yypt-4].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Switch, Expr: yyS[yypt-2].expr, Body: yyS[yypt-0].stmt}
		}
	case 84:
		//line cc.y:626
		{
			yyVAL.span = span(yyS[yypt-4].span, yyS[yypt-0].span)
			yyVAL.stmt = &Stmt{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: While, Expr: yyS[yypt-2].expr, Body: yyS[yypt-0].stmt}
		}
	case 85:
		//line cc.y:633
		{
			yyVAL.span = Span{}
			yyVAL.abdecor = func(t *Type) *Type { return t }
		}
	case 86:
		//line cc.y:638
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			_, q, _ := splitTypeWords(yyS[yypt-1].strs)
			abdecor := yyS[yypt-0].abdecor
			yyVAL.abdecor = func(t *Type) *Type {
				return abdecor(&Type{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Kind: Ptr, Base: t, Qual: q})
			}
		}
	case 87:
		//line cc.y:647
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.abdecor = yyS[yypt-0].abdecor
		}
	case 88:
		//line cc.y:654
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			abdecor := yyS[yypt-3].abdecor
			decls := yyS[yypt-1].decls
			span := yyVAL.span
			for _, decl := range decls {
				t := decl.Type
				if t != nil {
					if t.Kind == TypedefType && t.Base != nil {
						t = t.Base
					}
					if t.Kind == Array {
						if t.Width == nil {
							t = t.Base
						}
						decl.Type = &Type{Kind: Ptr, Base: t}
					}
				}
			}
			yyVAL.abdecor = func(t *Type) *Type {
				return abdecor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Func, Base: t, Decls: decls})
			}
		}
	case 89:
		//line cc.y:678
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			abdecor := yyS[yypt-3].abdecor
			span := yyVAL.span
			expr := yyS[yypt-1].expr
			yyVAL.abdecor = func(t *Type) *Type {
				return abdecor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Array, Base: t, Width: expr})
			}

		}
	case 90:
		//line cc.y:689
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.abdecor = yyS[yypt-1].abdecor
		}
	case 91:
		//line cc.y:697
		{
			yyVAL.span = yyS[yypt-0].span
			name := yyS[yypt-0].str
			yyVAL.decor = func(t *Type) (*Type, string) { return t, name }
		}
	case 92:
		//line cc.y:703
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			_, q, _ := splitTypeWords(yyS[yypt-1].strs)
			decor := yyS[yypt-0].decor
			span := yyVAL.span
			yyVAL.decor = func(t *Type) (*Type, string) {
				return decor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Ptr, Base: t, Qual: q})
			}
		}
	case 93:
		//line cc.y:713
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.decor = yyS[yypt-1].decor
		}
	case 94:
		//line cc.y:718
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			decor := yyS[yypt-3].decor
			decls := yyS[yypt-1].decls
			span := yyVAL.span
			yyVAL.decor = func(t *Type) (*Type, string) {
				return decor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Func, Base: t, Decls: decls})
			}
		}
	case 95:
		//line cc.y:728
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			decor := yyS[yypt-3].decor
			span := yyVAL.span
			expr := yyS[yypt-1].expr
			yyVAL.decor = func(t *Type) (*Type, string) {
				return decor(&Type{SyntaxInfo: SyntaxInfo{Span: span}, Kind: Array, Base: t, Width: expr})
			}
		}
	case 96:
		//line cc.y:741
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decl = &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: yyS[yypt-0].str}
		}
	case 97:
		//line cc.y:746
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.decl = &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Type: yyS[yypt-0].abdecor(yyS[yypt-1].typ)}
		}
	case 98:
		//line cc.y:751
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			typ, name := yyS[yypt-0].decor(yyS[yypt-1].typ)
			yyVAL.decl = &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: name, Type: typ}
		}
	case 99:
		//line cc.y:757
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decl = &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: "..."}
		}
	case 100:
		//line cc.y:765
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.idec = idecor{yyS[yypt-0].decor, nil}
		}
	case 101:
		//line cc.y:770
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.idec = idecor{yyS[yypt-2].decor, yyS[yypt-0].init}
		}
	case 102:
		//line cc.y:778
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 103:
		//line cc.y:783
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 104:
		//line cc.y:788
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 105:
		//line cc.y:793
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 106:
		//line cc.y:798
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 107:
		//line cc.y:803
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 108:
		//line cc.y:811
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 109:
		//line cc.y:816
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 110:
		//line cc.y:824
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 111:
		//line cc.y:829
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 112:
		//line cc.y:834
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 113:
		//line cc.y:839
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 114:
		//line cc.y:844
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 115:
		//line cc.y:849
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 116:
		//line cc.y:854
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 117:
		//line cc.y:859
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 118:
		//line cc.y:864
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 119:
		//line cc.y:871
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 120:
		//line cc.y:876
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 121:
		//line cc.y:883
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 122:
		//line cc.y:888
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 123:
		//line cc.y:896
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.typ = yyS[yypt-0].typ
			if yyVAL.typ == nil {
				yyVAL.typ = &Type{Kind: TypedefType, Name: yyS[yypt-0].str}
			}
		}
	case 124:
		//line cc.y:912
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.tc.c, yyVAL.tc.q, yyVAL.tc.t = splitTypeWords(append(yyS[yypt-0].strs, "int"))
		}
	case 125:
		//line cc.y:917
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.tc.c, yyVAL.tc.q, _ = splitTypeWords(append(yyS[yypt-2].strs, yyS[yypt-0].strs...))
			yyVAL.tc.t = yyS[yypt-1].typ
		}
	case 126:
		//line cc.y:923
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyS[yypt-2].strs = append(yyS[yypt-2].strs, yyS[yypt-1].str)
			yyS[yypt-2].strs = append(yyS[yypt-2].strs, yyS[yypt-0].strs...)
			yyVAL.tc.c, yyVAL.tc.q, yyVAL.tc.t = splitTypeWords(yyS[yypt-2].strs)
		}
	case 127:
		//line cc.y:930
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.tc.c, yyVAL.tc.q, _ = splitTypeWords(yyS[yypt-0].strs)
			yyVAL.tc.t = yyS[yypt-1].typ
		}
	case 128:
		//line cc.y:936
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			var ts []string
			ts = append(ts, yyS[yypt-1].str)
			ts = append(ts, yyS[yypt-0].strs...)
			yyVAL.tc.c, yyVAL.tc.q, yyVAL.tc.t = splitTypeWords(ts)
		}
	case 129:
		//line cc.y:947
		{
			yyVAL.span = yyS[yypt-0].span
			if yyS[yypt-0].tc.c != 0 {
				yylex.(*lexer).Errorf("%v not allowed here", yyS[yypt-0].tc.c)
			}
			if yyS[yypt-0].tc.q != 0 && yyS[yypt-0].tc.q != Const && yyS[yypt-0].tc.q != Volatile {
				yylex.(*lexer).Errorf("%v ignored here (TODO)?", yyS[yypt-0].tc.q)
			}
			yyVAL.typ = yyS[yypt-0].tc.t
		}
	case 130:
		//line cc.y:960
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.typ = yyS[yypt-0].abdecor(yyS[yypt-1].typ)
		}
	case 131:
		//line cc.y:968
		{
			lx := yylex.(*lexer)
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			// TODO: use $1.q
			yyVAL.decls = nil
			for _, idec := range yyS[yypt-1].idecs {
				typ, name := idec.d(yyS[yypt-2].tc.t)
				d := &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: name, Type: typ, Storage: yyS[yypt-2].tc.c, Init: idec.i}
				lx.pushDecl(d)
				yyVAL.decls = append(yyVAL.decls, d)
			}
			if yyS[yypt-1].idecs == nil {
				d := &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: "", Type: yyS[yypt-2].tc.t, Storage: yyS[yypt-2].tc.c}
				lx.pushDecl(d)
				yyVAL.decls = append(yyVAL.decls, d)
			}
		}
	case 132:
		//line cc.y:988
		{
			lx := yylex.(*lexer)
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			// TODO: use $1.q
			yyVAL.decls = nil
			for _, idec := range yyS[yypt-1].idecs {
				typ, name := idec.d(yyS[yypt-2].tc.t)
				d := lx.lookupDecl(name)
				if d == nil {
					d = &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: name, Type: typ, Storage: yyS[yypt-2].tc.c, Init: idec.i}
					lx.pushDecl(d)
				} else {
					d.Span = yyVAL.span
					if idec.i != nil {
						d.Init = idec.i
					}
				}
				yyVAL.decls = append(yyVAL.decls, d)
			}
			if yyS[yypt-1].idecs == nil {
				d := &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: "", Type: yyS[yypt-2].tc.t, Storage: yyS[yypt-2].tc.c}
				lx.pushDecl(d)
				yyVAL.decls = append(yyVAL.decls, d)
			}
		}
	case 133:
		//line cc.y:1016
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decls = yyS[yypt-0].decls
		}
	case 134:
		//line cc.y:1021
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decls = []*Decl{yyS[yypt-0].decl}
		}
	case 135:
		//line cc.y:1026
		{
			yyVAL.decls = yyS[yypt-1].decls
		}
	case 136:
		//line cc.y:1032
		{
			lx := yylex.(*lexer)
			typ, name := yyS[yypt-1].decor(yyS[yypt-2].tc.t)
			if typ.Kind != Func {
				yylex.(*lexer).Errorf("invalid function definition")
				return 0
			}
			d := lx.lookupDecl(name)
			if d == nil {
				d = &Decl{Name: name, Type: typ, Storage: yyS[yypt-2].tc.c}
				lx.pushDecl(d)
			} else {
				d.Type = typ
			}
			yyVAL.decl = d
			lx.pushScope()
			for _, decl := range typ.Decls {
				lx.pushDecl(decl)
			}
		}
	case 137:
		//line cc.y:1053
		{
			yylex.(*lexer).popScope()
			yyVAL.span = span(yyS[yypt-4].span, yyS[yypt-0].span)
			yyVAL.decl = yyS[yypt-1].decl
			yyVAL.decl.Span = yyVAL.span
			if yyS[yypt-2].decls != nil {
				yylex.(*lexer).Errorf("cannot use pre-prototype definitions")
			}
			yyVAL.decl.Body = yyS[yypt-0].stmt
		}
	case 138:
		//line cc.y:1066
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 139:
		//line cc.y:1071
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 140:
		//line cc.y:1079
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.tk = Struct
		}
	case 141:
		//line cc.y:1084
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.tk = Union
		}
	case 142:
		//line cc.y:1091
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decor = yyS[yypt-0].decor
		}
	case 143:
		//line cc.y:1096
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			name := yyS[yypt-2].str
			expr := yyS[yypt-0].expr
			yyVAL.decor = func(t *Type) (*Type, string) {
				t.Width = expr
				return t, name
			}
		}
	case 144:
		//line cc.y:1108
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.decls = nil
			for _, decor := range yyS[yypt-1].decors {
				typ, name := decor(yyS[yypt-2].typ)
				yyVAL.decls = append(yyVAL.decls, &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: name, Type: typ})
			}
			if yyS[yypt-1].decors == nil {
				yyVAL.decls = append(yyVAL.decls, &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Type: yyS[yypt-2].typ})
			}
		}
	case 145:
		//line cc.y:1122
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.typ = yylex.(*lexer).pushType(&Type{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Kind: yyS[yypt-1].tk, Tag: yyS[yypt-0].str})
		}
	case 146:
		//line cc.y:1127
		{
			yyVAL.span = span(yyS[yypt-4].span, yyS[yypt-0].span)
			yyVAL.typ = yylex.(*lexer).pushType(&Type{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Kind: yyS[yypt-4].tk, Tag: yyS[yypt-3].str, Decls: yyS[yypt-1].decls})
		}
	case 147:
		//line cc.y:1134
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.prefix = &Prefix{Span: yyVAL.span, Dot: yyS[yypt-0].str}
		}
	case 148:
		//line cc.y:1141
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Arrow, Left: yyS[yypt-2].expr, Text: yyS[yypt-0].str}
		}
	case 149:
		//line cc.y:1146
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.expr = &Expr{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Op: Dot, Left: yyS[yypt-2].expr, Text: yyS[yypt-0].str}
		}
	case 150:
		//line cc.y:1154
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.typ = yylex.(*lexer).pushType(&Type{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Kind: Enum, Tag: yyS[yypt-0].str})
		}
	case 151:
		//line cc.y:1159
		{
			yyVAL.span = span(yyS[yypt-5].span, yyS[yypt-0].span)
			yyVAL.typ = yylex.(*lexer).pushType(&Type{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Kind: Enum, Tag: yyS[yypt-4].str, Decls: yyS[yypt-2].decls})
		}
	case 152:
		//line cc.y:1166
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			var x *Init
			if yyS[yypt-0].expr != nil {
				x = &Init{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Expr: yyS[yypt-0].expr}
			}
			yyVAL.decl = &Decl{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Name: yyS[yypt-1].str, Init: x}
			yylex.(*lexer).pushDecl(yyVAL.decl)
		}
	case 153:
		//line cc.y:1178
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 154:
		//line cc.y:1186
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.init = &Init{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Expr: yyS[yypt-0].expr}
		}
	case 155:
		//line cc.y:1191
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.init = &Init{SyntaxInfo: SyntaxInfo{Span: yyVAL.span}, Braced: yyS[yypt-0].inits}
		}
	case 156:
		//line cc.y:1198
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.inits = []*Init{}
		}
	case 157:
		//line cc.y:1203
		{
			yyVAL.span = span(yyS[yypt-3].span, yyS[yypt-0].span)
			yyVAL.inits = append(yyS[yypt-2].inits, yyS[yypt-1].init)
		}
	case 158:
		//line cc.y:1208
		{
			yyVAL.span = span(yyS[yypt-4].span, yyS[yypt-0].span)
			yyVAL.inits = append(yyS[yypt-3].inits, yyS[yypt-2].init)
		}
	case 159:
		//line cc.y:1214
		{
			yyVAL.span = Span{}
			yyVAL.inits = nil
		}
	case 160:
		//line cc.y:1219
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.inits = append(yyS[yypt-2].inits, yyS[yypt-1].init)
		}
	case 161:
		//line cc.y:1226
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.init = yyS[yypt-0].init
		}
	case 162:
		//line cc.y:1231
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.init = yyS[yypt-0].init
			yyVAL.init.Prefix = yyS[yypt-2].prefixes
		}
	case 163:
		//line cc.y:1239
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.prefix = &Prefix{Span: yyVAL.span, Index: yyS[yypt-1].expr}
		}
	case 164:
		//line cc.y:1245
		{
			yyVAL.span = Span{}
		}
	case 165:
		//line cc.y:1249
		{
			yyVAL.span = yyS[yypt-0].span
		}
	case 166:
		//line cc.y:1254
		{
			yyVAL.span = Span{}
		}
	case 167:
		//line cc.y:1258
		{
			yyVAL.span = yyS[yypt-0].span
		}
	case 168:
		//line cc.y:1267
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.prefixes = []*Prefix{yyS[yypt-0].prefix}
		}
	case 169:
		//line cc.y:1272
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.prefixes = append(yyS[yypt-1].prefixes, yyS[yypt-0].prefix)
		}
	case 170:
		//line cc.y:1278
		{
			yyVAL.span = Span{}
			yyVAL.str = ""
		}
	case 171:
		//line cc.y:1283
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.str = yyS[yypt-0].str
		}
	case 172:
		//line cc.y:1289
		{
			yyVAL.span = Span{}
			yyVAL.expr = nil
		}
	case 173:
		//line cc.y:1294
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 174:
		//line cc.y:1300
		{
			yyVAL.span = Span{}
			yyVAL.expr = nil
		}
	case 175:
		//line cc.y:1305
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 176:
		//line cc.y:1312
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.exprs = []*Expr{yyS[yypt-0].expr}
		}
	case 177:
		//line cc.y:1317
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.exprs = append(yyS[yypt-2].exprs, yyS[yypt-0].expr)
		}
	case 178:
		//line cc.y:1323
		{
			yyVAL.span = Span{}
			yyVAL.exprs = nil
		}
	case 179:
		//line cc.y:1328
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.exprs = yyS[yypt-0].exprs
		}
	case 180:
		//line cc.y:1334
		{
			yyVAL.span = Span{}
			yyVAL.decls = nil
		}
	case 181:
		//line cc.y:1339
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.decls = append(yyS[yypt-1].decls, yyS[yypt-0].decls...)
		}
	case 182:
		//line cc.y:1345
		{
			yyVAL.span = Span{}
			yyVAL.labels = nil
		}
	case 183:
		//line cc.y:1350
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.labels = append(yyS[yypt-1].labels, yyS[yypt-0].label)
		}
	case 184:
		//line cc.y:1357
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decls = []*Decl{yyS[yypt-0].decl}
		}
	case 185:
		//line cc.y:1362
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.decls = append(yyS[yypt-2].decls, yyS[yypt-0].decl)
		}
	case 186:
		//line cc.y:1368
		{
			yyVAL.span = Span{}
			yyVAL.decls = nil
		}
	case 187:
		//line cc.y:1373
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decls = yyS[yypt-0].decls
		}
	case 188:
		//line cc.y:1380
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.idecs = []idecor{yyS[yypt-0].idec}
		}
	case 189:
		//line cc.y:1385
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.idecs = append(yyS[yypt-2].idecs, yyS[yypt-0].idec)
		}
	case 190:
		//line cc.y:1391
		{
			yyVAL.span = Span{}
			yyVAL.idecs = nil
		}
	case 191:
		//line cc.y:1396
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.idecs = yyS[yypt-0].idecs
		}
	case 192:
		//line cc.y:1403
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.strs = []string{yyS[yypt-0].str}
		}
	case 193:
		//line cc.y:1408
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.strs = append(yyS[yypt-1].strs, yyS[yypt-0].str)
		}
	case 194:
		//line cc.y:1414
		{
			yyVAL.span = Span{}
			yyVAL.strs = nil
		}
	case 195:
		//line cc.y:1419
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.strs = yyS[yypt-0].strs
		}
	case 196:
		//line cc.y:1426
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.strs = []string{yyS[yypt-0].str}
		}
	case 197:
		//line cc.y:1431
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.strs = append(yyS[yypt-1].strs, yyS[yypt-0].str)
		}
	case 198:
		//line cc.y:1437
		{
			yyVAL.span = Span{}
			yyVAL.strs = nil
		}
	case 199:
		//line cc.y:1442
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.strs = yyS[yypt-0].strs
		}
	case 200:
		//line cc.y:1449
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.strs = []string{yyS[yypt-0].str}
		}
	case 201:
		//line cc.y:1454
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.strs = append(yyS[yypt-1].strs, yyS[yypt-0].str)
		}
	case 202:
		//line cc.y:1460
		{
			yyVAL.span = Span{}
			yyVAL.strs = nil
		}
	case 203:
		//line cc.y:1465
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.strs = yyS[yypt-0].strs
		}
	case 204:
		//line cc.y:1472
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decors = nil
			yyVAL.decors = append(yyVAL.decors, yyS[yypt-0].decor)
		}
	case 205:
		//line cc.y:1478
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.decors = append(yyS[yypt-2].decors, yyS[yypt-0].decor)
		}
	case 206:
		//line cc.y:1484
		{
			yyVAL.span = Span{}
			yyVAL.decors = nil
		}
	case 207:
		//line cc.y:1489
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decors = yyS[yypt-0].decors
		}
	case 208:
		//line cc.y:1496
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decls = yyS[yypt-0].decls
		}
	case 209:
		//line cc.y:1501
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.decls = append(yyS[yypt-1].decls, yyS[yypt-0].decls...)
		}
	case 210:
		//line cc.y:1507
		{
			yyVAL.span = Span{}
			yyVAL.expr = nil
		}
	case 211:
		//line cc.y:1512
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 212:
		//line cc.y:1519
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.decls = []*Decl{yyS[yypt-0].decl}
		}
	case 213:
		//line cc.y:1524
		{
			yyVAL.span = span(yyS[yypt-2].span, yyS[yypt-0].span)
			yyVAL.decls = append(yyS[yypt-2].decls, yyS[yypt-0].decl)
		}
	case 214:
		//line cc.y:1531
		{
			yyVAL.span = yyS[yypt-0].span
			yyVAL.strs = []string{yyS[yypt-0].str}
		}
	case 215:
		//line cc.y:1536
		{
			yyVAL.span = span(yyS[yypt-1].span, yyS[yypt-0].span)
			yyVAL.strs = append(yyS[yypt-1].strs, yyS[yypt-0].str)
		}
	}
	goto yystack /* stack new state and value */
}
