// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloudprint

// defaultPPD is the default PPD (capabilities) for a new printer.
var defaultPPD = `*PPD-Adobe: "4.3"
*FormatVersion:	"4.3"
*FileVersion:	"1.1"
*LanguageVersion: English
*LanguageEncoding: ISOLatin1
*PCFileName:	"GOCLOUD.PPD"
*Manufacturer:	"Generic"
*Product:	"Go Virtual Printer"
*ModelName:     "Generic Cloud Printer"
*ShortNickName: "Generic Cloud Printer"
*NickName:      "Generic Cloud Printer"
*PSVersion:	"(2017.000) 0"
*LanguageLevel:	"2"
*ColorDevice:	True
*DefaultColorSpace: RGB
*FileSystem:	False
*Throughput:	"8"
*LandscapeOrientation: Plus90
*TTRasterizer:	Type42
*DefaultResolution: 600dpi

*HWMargins: 0 0 0 0
*VariablePaperSize: True
*MaxMediaWidth: 100000
*MaxMediaHeight: 100000
*NonUIOrderDependency: 100 AnySetup *CustomPageSize
*CustomPageSize True: "pop pop pop
<</PageSize [ 5 -2 roll ] /ImagingBBox null>>setpagedevice"
*End
*ParamCustomPageSize Width: 1 points 36 100000
*ParamCustomPageSize Height: 2 points 36 100000
*ParamCustomPageSize Orientation: 3 int 0 3
*ParamCustomPageSize WidthOffset: 4 points 0 0
*ParamCustomPageSize HeightOffset: 5 points 0 0

*OpenGroup: General/General

*OpenUI *PageSize/Page Size: PickOne
*OrderDependency: 100 AnySetup *PageSize
*DefaultPageSize: Letter
*PageSize 11x14/11x14: "<</PageSize[792 1008]/ImagingBBox null>>setpagedevice"
*PageSize 11x17/11x17: "<</PageSize[792 1224]/ImagingBBox null>>setpagedevice"
*PageSize 13x19/13x19: "<</PageSize[936 1368]/ImagingBBox null>>setpagedevice"
*PageSize 16x20/16x20: "<</PageSize[1152 1440]/ImagingBBox null>>setpagedevice"
*PageSize 16x24/16x24: "<</PageSize[1152 1728]/ImagingBBox null>>setpagedevice"
*PageSize 2A/2A: "<</PageSize[3370 4768]/ImagingBBox null>>setpagedevice"
*PageSize 4A/4A: "<</PageSize[4768 6749]/ImagingBBox null>>setpagedevice"
*PageSize 8x10/8x10: "<</PageSize[576 720]/ImagingBBox null>>setpagedevice"
*PageSize 8x12/8x12: "<</PageSize[576 864]/ImagingBBox null>>setpagedevice"
*PageSize A0/A0: "<</PageSize[2384 3370]/ImagingBBox null>>setpagedevice"
*PageSize A1/A1: "<</PageSize[1684 2384]/ImagingBBox null>>setpagedevice"
*PageSize A2/A2: "<</PageSize[1191 1684]/ImagingBBox null>>setpagedevice"
*PageSize A3/A3: "<</PageSize[842 1191]/ImagingBBox null>>setpagedevice"
*PageSize A4/A4: "<</PageSize[595 842]/ImagingBBox null>>setpagedevice"
*PageSize A5/A5: "<</PageSize[421 595]/ImagingBBox null>>setpagedevice"
*PageSize AnsiA/ANSI A: "<</PageSize[612 792]/ImagingBBox null>>setpagedevice"
*PageSize AnsiB/ANSI B: "<</PageSize[792 1224]/ImagingBBox null>>setpagedevice"
*PageSize AnsiC/ANSI C: "<</PageSize[1224 1584]/ImagingBBox null>>setpagedevice"
*PageSize AnsiD/ANSI D: "<</PageSize[1584 2448]/ImagingBBox null>>setpagedevice"
*PageSize AnsiE/ANSI E: "<</PageSize[2448 3168]/ImagingBBox null>>setpagedevice"
*PageSize ArchA/Arch A: "<</PageSize[648 864]/ImagingBBox null>>setpagedevice"
*PageSize ArchB/Arch B: "<</PageSize[864 1296]/ImagingBBox null>>setpagedevice"
*PageSize ArchC/Arch C: "<</PageSize[1296 1728]/ImagingBBox null>>setpagedevice"
*PageSize ArchD/Arch D: "<</PageSize[1728 2592]/ImagingBBox null>>setpagedevice"
*PageSize ArchE/Arch E: "<</PageSize[2592 3456]/ImagingBBox null>>setpagedevice"
*PageSize C0/C0: "<</PageSize[2599 3676]/ImagingBBox null>>setpagedevice"
*PageSize C1/C1: "<</PageSize[1836 2599]/ImagingBBox null>>setpagedevice"
*PageSize C2/C2: "<</PageSize[1298 1836]/ImagingBBox null>>setpagedevice"
*PageSize C3/C3: "<</PageSize[918 1298]/ImagingBBox null>>setpagedevice"
*PageSize C4/C4: "<</PageSize[649 918]/ImagingBBox null>>setpagedevice"
*PageSize C5/C5: "<</PageSize[459 649]/ImagingBBox null>>setpagedevice"
*PageSize Env10/Envelope #10: "<</PageSize[297 684]/ImagingBBox null>>setpagedevice"
*PageSize EnvC5/Envelope C5: "<</PageSize[459 649]/ImagingBBox null>>setpagedevice"
*PageSize EnvDL/Envelope DL: "<</PageSize[312 624]/ImagingBBox null>>setpagedevice"
*PageSize EnvMonarch/Envelope Monarch: "<</PageSize[279 540]/ImagingBBox null>>setpagedevice"
*PageSize Executive/Executive: "<</PageSize[522 756]/ImagingBBox null>>setpagedevice"
*PageSize ISOB0/B0 (ISO): "<</PageSize[2834 4008]/ImagingBBox null>>setpagedevice"
*PageSize ISOB1/B1 (ISO): "<</PageSize[2004 2834]/ImagingBBox null>>setpagedevice"
*PageSize ISOB2/B2 (ISO): "<</PageSize[1417 2004]/ImagingBBox null>>setpagedevice"
*PageSize ISOB3/B3 (ISO): "<</PageSize[1000 1417]/ImagingBBox null>>setpagedevice"
*PageSize ISOB4/B4 (ISO): "<</PageSize[708 1000]/ImagingBBox null>>setpagedevice"
*PageSize ISOB5/B5 (ISO): "<</PageSize[498 708]/ImagingBBox null>>setpagedevice"
*PageSize JISB0/B0 (JIS): "<</PageSize[2919 4127]/ImagingBBox null>>setpagedevice"
*PageSize JISB1/B1 (JIS): "<</PageSize[2063 2919]/ImagingBBox null>>setpagedevice"
*PageSize JISB2/B2 (JIS): "<</PageSize[1459 2063]/ImagingBBox null>>setpagedevice"
*PageSize JISB3/B3 (JIS): "<</PageSize[1029 1459]/ImagingBBox null>>setpagedevice"
*PageSize JISB4/B4 (JIS): "<</PageSize[727 1029]/ImagingBBox null>>setpagedevice"
*PageSize JISB5/B5 (JIS): "<</PageSize[518 727]/ImagingBBox null>>setpagedevice"
*PageSize Ledger/Ledger: "<</PageSize[1224 792]/ImagingBBox null>>setpagedevice"
*PageSize Legal/US Legal: "<</PageSize[612 1008]/ImagingBBox null>>setpagedevice"
*PageSize Letter/US Letter: "<</PageSize[612 792]/ImagingBBox null>>setpagedevice"
*PageSize RA0/RA0: "<</PageSize[2437 3458]/ImagingBBox null>>setpagedevice"
*PageSize RA1/RA1: "<</PageSize[1729 2437]/ImagingBBox null>>setpagedevice"
*PageSize RA2/RA2: "<</PageSize[1218 1729]/ImagingBBox null>>setpagedevice"
*PageSize RA3/RA3: "<</PageSize[864 1218]/ImagingBBox null>>setpagedevice"
*PageSize RA4/RA4: "<</PageSize[609 864]/ImagingBBox null>>setpagedevice"
*PageSize SRA0/SRA0: "<</PageSize[2551 3628]/ImagingBBox null>>setpagedevice"
*PageSize SRA1/SRA1: "<</PageSize[1814 2551]/ImagingBBox null>>setpagedevice"
*PageSize SRA2/SRA2: "<</PageSize[1275 1814]/ImagingBBox null>>setpagedevice"
*PageSize SRA3/SRA3: "<</PageSize[907 1275]/ImagingBBox null>>setpagedevice"
*PageSize SRA4/SRA4: "<</PageSize[637 907]/ImagingBBox null>>setpagedevice"
*PageSize SuperA/Super A: "<</PageSize[644 1008]/ImagingBBox null>>setpagedevice"
*PageSize SuperB/Super B: "<</PageSize[936 1368]/ImagingBBox null>>setpagedevice"
*PageSize TabloidExtra/Tabloid Extra: "<</PageSize[864 1296]/ImagingBBox null>>setpagedevice"
*PageSize Tabloid/Tabloid: "<</PageSize[792 1224]/ImagingBBox null>>setpagedevice"
*CloseUI: *PageSize

*OpenUI *PageRegion: PickOne
*OrderDependency: 100 AnySetup *PageRegion
*DefaultPageRegion: Letter
*PageRegion 11x14/11x14: "<</PageSize[792 1008]/ImagingBBox null>>setpagedevice"
*PageRegion 11x17/11x17: "<</PageSize[792 1224]/ImagingBBox null>>setpagedevice"
*PageRegion 13x19/13x19: "<</PageSize[936 1368]/ImagingBBox null>>setpagedevice"
*PageRegion 16x20/16x20: "<</PageSize[1152 1440]/ImagingBBox null>>setpagedevice"
*PageRegion 16x24/16x24: "<</PageSize[1152 1728]/ImagingBBox null>>setpagedevice"
*PageRegion 2A/2A: "<</PageSize[3370 4768]/ImagingBBox null>>setpagedevice"
*PageRegion 4A/4A: "<</PageSize[4768 6749]/ImagingBBox null>>setpagedevice"
*PageRegion 8x10/8x10: "<</PageSize[576 720]/ImagingBBox null>>setpagedevice"
*PageRegion 8x12/8x12: "<</PageSize[576 864]/ImagingBBox null>>setpagedevice"
*PageRegion A0/A0: "<</PageSize[2384 3370]/ImagingBBox null>>setpagedevice"
*PageRegion A1/A1: "<</PageSize[1684 2384]/ImagingBBox null>>setpagedevice"
*PageRegion A2/A2: "<</PageSize[1191 1684]/ImagingBBox null>>setpagedevice"
*PageRegion A3/A3: "<</PageSize[842 1191]/ImagingBBox null>>setpagedevice"
*PageRegion A4/A4: "<</PageSize[595 842]/ImagingBBox null>>setpagedevice"
*PageRegion A5/A5: "<</PageSize[421 595]/ImagingBBox null>>setpagedevice"
*PageRegion AnsiA/ANSI A: "<</PageSize[612 792]/ImagingBBox null>>setpagedevice"
*PageRegion AnsiB/ANSI B: "<</PageSize[792 1224]/ImagingBBox null>>setpagedevice"
*PageRegion AnsiC/ANSI C: "<</PageSize[1224 1584]/ImagingBBox null>>setpagedevice"
*PageRegion AnsiD/ANSI D: "<</PageSize[1584 2448]/ImagingBBox null>>setpagedevice"
*PageRegion AnsiE/ANSI E: "<</PageSize[2448 3168]/ImagingBBox null>>setpagedevice"
*PageRegion ArchA/Arch A: "<</PageSize[648 864]/ImagingBBox null>>setpagedevice"
*PageRegion ArchB/Arch B: "<</PageSize[864 1296]/ImagingBBox null>>setpagedevice"
*PageRegion ArchC/Arch C: "<</PageSize[1296 1728]/ImagingBBox null>>setpagedevice"
*PageRegion ArchD/Arch D: "<</PageSize[1728 2592]/ImagingBBox null>>setpagedevice"
*PageRegion ArchE/Arch E: "<</PageSize[2592 3456]/ImagingBBox null>>setpagedevice"
*PageRegion C0/C0: "<</PageSize[2599 3676]/ImagingBBox null>>setpagedevice"
*PageRegion C1/C1: "<</PageSize[1836 2599]/ImagingBBox null>>setpagedevice"
*PageRegion C2/C2: "<</PageSize[1298 1836]/ImagingBBox null>>setpagedevice"
*PageRegion C3/C3: "<</PageSize[918 1298]/ImagingBBox null>>setpagedevice"
*PageRegion C4/C4: "<</PageSize[649 918]/ImagingBBox null>>setpagedevice"
*PageRegion C5/C5: "<</PageSize[459 649]/ImagingBBox null>>setpagedevice"
*PageRegion Env10/Envelope #10: "<</PageSize[297 684]/ImagingBBox null>>setpagedevice"
*PageRegion EnvC5/Envelope C5: "<</PageSize[459 649]/ImagingBBox null>>setpagedevice"
*PageRegion EnvDL/Envelope DL: "<</PageSize[312 624]/ImagingBBox null>>setpagedevice"
*PageRegion EnvMonarch/Envelope Monarch: "<</PageSize[279 540]/ImagingBBox null>>setpagedevice"
*PageRegion Executive/Executive: "<</PageSize[522 756]/ImagingBBox null>>setpagedevice"
*PageRegion ISOB0/B0 (ISO): "<</PageSize[2834 4008]/ImagingBBox null>>setpagedevice"
*PageRegion ISOB1/B1 (ISO): "<</PageSize[2004 2834]/ImagingBBox null>>setpagedevice"
*PageRegion ISOB2/B2 (ISO): "<</PageSize[1417 2004]/ImagingBBox null>>setpagedevice"
*PageRegion ISOB3/B3 (ISO): "<</PageSize[1000 1417]/ImagingBBox null>>setpagedevice"
*PageRegion ISOB4/B4 (ISO): "<</PageSize[708 1000]/ImagingBBox null>>setpagedevice"
*PageRegion ISOB5/B5 (ISO): "<</PageSize[498 708]/ImagingBBox null>>setpagedevice"
*PageRegion JISB0/B0 (JIS): "<</PageSize[2919 4127]/ImagingBBox null>>setpagedevice"
*PageRegion JISB1/B1 (JIS): "<</PageSize[2063 2919]/ImagingBBox null>>setpagedevice"
*PageRegion JISB2/B2 (JIS): "<</PageSize[1459 2063]/ImagingBBox null>>setpagedevice"
*PageRegion JISB3/B3 (JIS): "<</PageSize[1029 1459]/ImagingBBox null>>setpagedevice"
*PageRegion JISB4/B4 (JIS): "<</PageSize[727 1029]/ImagingBBox null>>setpagedevice"
*PageRegion JISB5/B5 (JIS): "<</PageSize[518 727]/ImagingBBox null>>setpagedevice"
*PageRegion Ledger/Ledger: "<</PageSize[1224 792]/ImagingBBox null>>setpagedevice"
*PageRegion Legal/US Legal: "<</PageSize[612 1008]/ImagingBBox null>>setpagedevice"
*PageRegion Letter/US Letter: "<</PageSize[612 792]/ImagingBBox null>>setpagedevice"
*PageRegion RA0/RA0: "<</PageSize[2437 3458]/ImagingBBox null>>setpagedevice"
*PageRegion RA1/RA1: "<</PageSize[1729 2437]/ImagingBBox null>>setpagedevice"
*PageRegion RA2/RA2: "<</PageSize[1218 1729]/ImagingBBox null>>setpagedevice"
*PageRegion RA3/RA3: "<</PageSize[864 1218]/ImagingBBox null>>setpagedevice"
*PageRegion RA4/RA4: "<</PageSize[609 864]/ImagingBBox null>>setpagedevice"
*PageRegion SRA0/SRA0: "<</PageSize[2551 3628]/ImagingBBox null>>setpagedevice"
*PageRegion SRA1/SRA1: "<</PageSize[1814 2551]/ImagingBBox null>>setpagedevice"
*PageRegion SRA2/SRA2: "<</PageSize[1275 1814]/ImagingBBox null>>setpagedevice"
*PageRegion SRA3/SRA3: "<</PageSize[907 1275]/ImagingBBox null>>setpagedevice"
*PageRegion SRA4/SRA4: "<</PageSize[637 907]/ImagingBBox null>>setpagedevice"
*PageRegion SuperA/Super A: "<</PageSize[644 1008]/ImagingBBox null>>setpagedevice"
*PageRegion SuperB/Super B: "<</PageSize[936 1368]/ImagingBBox null>>setpagedevice"
*PageRegion TabloidExtra/Tabloid Extra: "<</PageSize[864 1296]/ImagingBBox null>>setpagedevice"
*PageRegion Tabloid/Tabloid: "<</PageSize[792 1224]/ImagingBBox null>>setpagedevice"
*CloseUI: *PageRegion

*DefaultImageableArea: Letter
*ImageableArea 11x14/11x14: "0 0 792 1008"
*ImageableArea 11x17/11x17: "0 0 792 1224"
*ImageableArea 13x19/13x19: "0 0 936 1368"
*ImageableArea 16x20/16x20: "0 0 1152 1440"
*ImageableArea 16x24/16x24: "0 0 1152 1728"
*ImageableArea 2A/2A: "0 0 3370 4768"
*ImageableArea 4A/4A: "0 0 4768 6749"
*ImageableArea 8x10/8x10: "0 0 576 720"
*ImageableArea 8x12/8x12: "0 0 576 864"
*ImageableArea A0/A0: "0 0 2384 3370"
*ImageableArea A1/A1: "0 0 1684 2384"
*ImageableArea A2/A2: "0 0 1191 1684"
*ImageableArea A3/A3: "0 0 842 1191"
*ImageableArea A4/A4: "0 0 595 842"
*ImageableArea A5/A5: "0 0 421 595"
*ImageableArea AnsiA/ANSI A: "0 0 612 792"
*ImageableArea AnsiB/ANSI B: "0 0 792 1224"
*ImageableArea AnsiC/ANSI C: "0 0 1224 1584"
*ImageableArea AnsiD/ANSI D: "0 0 1584 2448"
*ImageableArea AnsiE/ANSI E: "0 0 2448 3168"
*ImageableArea ArchA/Arch A: "0 0 648 864"
*ImageableArea ArchB/Arch B: "0 0 864 1296"
*ImageableArea ArchC/Arch C: "0 0 1296 1728"
*ImageableArea ArchD/Arch D: "0 0 1728 2592"
*ImageableArea ArchE/Arch E: "0 0 2592 3456"
*ImageableArea C0/C0: "0 0 2599 3676"
*ImageableArea C1/C1: "0 0 1836 2599"
*ImageableArea C2/C2: "0 0 1298 1836"
*ImageableArea C3/C3: "0 0 918 1298"
*ImageableArea C4/C4: "0 0 649 918"
*ImageableArea C5/C5: "0 0 459 649"
*ImageableArea Env10/Envelope #10: "0 0 297 684"
*ImageableArea EnvC5/Envelope C5: "0 0 459 649"
*ImageableArea EnvDL/Envelope DL: "0 0 312 624"
*ImageableArea EnvMonarch/Envelope Monarch: "0 0 279 540"
*ImageableArea Executive/Executive: "0 0 522 756"
*ImageableArea ISOB0/B0 (ISO): "0 0 2834 4008"
*ImageableArea ISOB1/B1 (ISO): "0 0 2004 2834"
*ImageableArea ISOB2/B2 (ISO): "0 0 1417 2004"
*ImageableArea ISOB3/B3 (ISO): "0 0 1000 1417"
*ImageableArea ISOB4/B4 (ISO): "0 0 708 1000"
*ImageableArea ISOB5/B5 (ISO): "0 0 498 708"
*ImageableArea JISB0/B0 (JIS): "0 0 2919 4127"
*ImageableArea JISB1/B1 (JIS): "0 0 2063 2919"
*ImageableArea JISB2/B2 (JIS): "0 0 1459 2063"
*ImageableArea JISB3/B3 (JIS): "0 0 1029 1459"
*ImageableArea JISB4/B4 (JIS): "0 0 727 1029"
*ImageableArea JISB5/B5 (JIS): "0 0 518 727"
*ImageableArea Ledger/Ledger: "0 0 1224 792"
*ImageableArea Legal/US Legal: "0 0 612 1008"
*ImageableArea Letter/US Letter: "0 0 612 792"
*ImageableArea RA0/RA0: "0 0 2437 3458"
*ImageableArea RA1/RA1: "0 0 1729 2437"
*ImageableArea RA2/RA2: "0 0 1218 1729"
*ImageableArea RA3/RA3: "0 0 864 1218"
*ImageableArea RA4/RA4: "0 0 609 864"
*ImageableArea SRA0/SRA0: "0 0 2551 3628"
*ImageableArea SRA1/SRA1: "0 0 1814 2551"
*ImageableArea SRA2/SRA2: "0 0 1275 1814"
*ImageableArea SRA3/SRA3: "0 0 907 1275"
*ImageableArea SRA4/SRA4: "0 0 637 907"
*ImageableArea SuperA/Super A: "0 0 644 1008"
*ImageableArea SuperB/Super B: "0 0 936 1368"
*ImageableArea TabloidExtra/Tabloid Extra: "0 0 864 1296"
*ImageableArea Tabloid/Tabloid: "0 0 792 1224"

*DefaultPaperDimension: Letter
*PaperDimension 11x14/11x14: "792 1008"
*PaperDimension 11x17/11x17: "792 1224"
*PaperDimension 13x19/13x19: "936 1368"
*PaperDimension 16x20/16x20: "1152 1440"
*PaperDimension 16x24/16x24: "1152 1728"
*PaperDimension 2A/2A: "3370 4768"
*PaperDimension 4A/4A: "4768 6749"
*PaperDimension 8x10/8x10: "576 720"
*PaperDimension 8x12/8x12: "576 864"
*PaperDimension A0/A0: "2384 3370"
*PaperDimension A1/A1: "1684 2384"
*PaperDimension A2/A2: "1191 1684"
*PaperDimension A3/A3: "842 1191"
*PaperDimension A4/A4: "595 842"
*PaperDimension A5/A5: "421 595"
*PaperDimension AnsiA/ANSI A: "612 792"
*PaperDimension AnsiB/ANSI B: "792 1224"
*PaperDimension AnsiC/ANSI C: "1224 1584"
*PaperDimension AnsiD/ANSI D: "1584 2448"
*PaperDimension AnsiE/ANSI E: "2448 3168"
*PaperDimension ArchA/Arch A: "648 864"
*PaperDimension ArchB/Arch B: "864 1296"
*PaperDimension ArchC/Arch C: "1296 1728"
*PaperDimension ArchD/Arch D: "1728 2592"
*PaperDimension ArchE/Arch E: "2592 3456"
*PaperDimension C0/C0: "2599 3676"
*PaperDimension C1/C1: "1836 2599"
*PaperDimension C2/C2: "1298 1836"
*PaperDimension C3/C3: "918 1298"
*PaperDimension C4/C4: "649 918"
*PaperDimension C5/C5: "459 649"
*PaperDimension Env10/Envelope #10: "297 684"
*PaperDimension EnvC5/Envelope C5: "459 649"
*PaperDimension EnvDL/Envelope DL: "312 624"
*PaperDimension EnvMonarch/Envelope Monarch: "279 540"
*PaperDimension Executive/Executive: "522 756"
*PaperDimension ISOB0/B0 (ISO): "2834 4008"
*PaperDimension ISOB1/B1 (ISO): "2004 2834"
*PaperDimension ISOB2/B2 (ISO): "1417 2004"
*PaperDimension ISOB3/B3 (ISO): "1000 1417"
*PaperDimension ISOB4/B4 (ISO): "708 1000"
*PaperDimension ISOB5/B5 (ISO): "498 708"
*PaperDimension JISB0/B0 (JIS): "2919 4127"
*PaperDimension JISB1/B1 (JIS): "2063 2919"
*PaperDimension JISB2/B2 (JIS): "1459 2063"
*PaperDimension JISB3/B3 (JIS): "1029 1459"
*PaperDimension JISB4/B4 (JIS): "727 1029"
*PaperDimension JISB5/B5 (JIS): "518 727"
*PaperDimension Ledger/Ledger: "1224 792"
*PaperDimension Legal/US Legal: "612 1008"
*PaperDimension Letter/US Letter: "612 792"
*PaperDimension RA0/RA0: "2437 3458"
*PaperDimension RA1/RA1: "1729 2437"
*PaperDimension RA2/RA2: "1218 1729"
*PaperDimension RA3/RA3: "864 1218"
*PaperDimension RA4/RA4: "609 864"
*PaperDimension SRA0/SRA0: "2551 3628"
*PaperDimension SRA1/SRA1: "1814 2551"
*PaperDimension SRA2/SRA2: "1275 1814"
*PaperDimension SRA3/SRA3: "907 1275"
*PaperDimension SRA4/SRA4: "637 907"
*PaperDimension SuperA/Super A: "644 1008"
*PaperDimension SuperB/Super B: "936 1368"
*PaperDimension TabloidExtra/Tabloid Extra: "864 1296"
*PaperDimension Tabloid/Tabloid: "792 1224"

*OpenUI *Resolution/Output Resolution: PickOne
*OrderDependency: 100 AnySetup *Resolution
*DefaultResolution: 300dpi
*Resolution 150dpi/150 DPI: "<</HWResolution[150 150]>>setpagedevice"
*Resolution 300dpi/300 DPI: "<</HWResolution[300 300]>>setpagedevice"
*Resolution 600dpi/600 DPI: "<</HWResolution[600 600]>>setpagedevice"
*Resolution 1200dpi/1200 DPI: "<</HWResolution[1200 1200]>>setpagedevice"
*Resolution 2400dpi/2400 DPI: "<</HWResolution[2400 2400]>>setpagedevice"
*CloseUI: *Resolution

*CloseGroup: General

*DefaultFont: Courier
*Font AvantGarde-Book: Standard "(001.006S)" Standard ROM
*Font AvantGarde-BookOblique: Standard "(001.006S)" Standard ROM
*Font AvantGarde-Demi: Standard "(001.007S)" Standard ROM
*Font AvantGarde-DemiOblique: Standard "(001.007S)" Standard ROM
*Font Bookman-Demi: Standard "(001.004S)" Standard ROM
*Font Bookman-DemiItalic: Standard "(001.004S)" Standard ROM
*Font Bookman-Light: Standard "(001.004S)" Standard ROM
*Font Bookman-LightItalic: Standard "(001.004S)" Standard ROM
*Font Courier: Standard "(002.004S)" Standard ROM
*Font Courier-Bold: Standard "(002.004S)" Standard ROM
*Font Courier-BoldOblique: Standard "(002.004S)" Standard ROM
*Font Courier-Oblique: Standard "(002.004S)" Standard ROM
*Font Helvetica: Standard "(001.006S)" Standard ROM
*Font Helvetica-Bold: Standard "(001.007S)" Standard ROM
*Font Helvetica-BoldOblique: Standard "(001.007S)" Standard ROM
*Font Helvetica-Narrow: Standard "(001.006S)" Standard ROM
*Font Helvetica-Narrow-Bold: Standard "(001.007S)" Standard ROM
*Font Helvetica-Narrow-BoldOblique: Standard "(001.007S)" Standard ROM
*Font Helvetica-Narrow-Oblique: Standard "(001.006S)" Standard ROM
*Font Helvetica-Oblique: Standard "(001.006S)" Standard ROM
*Font NewCenturySchlbk-Bold: Standard "(001.009S)" Standard ROM
*Font NewCenturySchlbk-BoldItalic: Standard "(001.007S)" Standard ROM
*Font NewCenturySchlbk-Italic: Standard "(001.006S)" Standard ROM
*Font NewCenturySchlbk-Roman: Standard "(001.007S)" Standard ROM
*Font Palatino-Bold: Standard "(001.005S)" Standard ROM
*Font Palatino-BoldItalic: Standard "(001.005S)" Standard ROM
*Font Palatino-Italic: Standard "(001.005S)" Standard ROM
*Font Palatino-Roman: Standard "(001.005S)" Standard ROM
*Font Symbol: Special "(001.007S)" Special ROM
*Font Times-Bold: Standard "(001.007S)" Standard ROM
*Font Times-BoldItalic: Standard "(001.009S)" Standard ROM
*Font Times-Italic: Standard "(001.007S)" Standard ROM
*Font Times-Roman: Standard "(001.007S)" Standard ROM
*Font ZapfChancery-MediumItalic: Standard "(001.007S)" Standard ROM
*Font ZapfDingbats: Special "(001.004S)" Standard ROM
`
