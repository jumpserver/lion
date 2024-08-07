package guacd

var RDPServerLayouts = map[string]string{
	"de-de-qwertz":    "de-de-qwertz",
	"de-ch-qwertz":    "de-ch-qwertz",
	"en-gb-qwerty":    "en-gb-qwerty",
	"en-us-qwerty":    "en-us-qwerty",
	"es-es-qwerty":    "es-es-qwerty",
	"es-latam-qwerty": "es-latam-qwerty",
	"failsafe":        "failsafe",
	"fr-be-azerty":    "fr-be-azerty",
	"fr-fr-azerty":    "fr-fr-azerty",
	"fr-ca-qwerty":    "fr-ca-qwerty",
	"fr-ch-qwertz":    "fr-ch-qwertz",
	"hu-hu-qwertz":    "hu-hu-qwertz",
	"it-it-qwerty":    "it-it-qwerty",
	"ja-jp-qwerty":    "ja-jp-qwerty",
	"no-no-qwerty":    "no-no-qwerty",
	"pl-pl-qwerty":    "pl-pl-qwerty",
	"pt-br-qwerty":    "pt-br-qwerty",
	"pt-pt-qwerty":    "pt-pt-qwerty",
	"ro-ro-qwerty":    "ro-ro-qwerty",
	"sv-se-qwerty":    "sv-se-qwerty",
	"da-dk-qwerty":    "da-dk-qwerty",
	"tr-tr-qwerty":    "tr-tr-qwerty",
}

/*
https://github.com/apache/guacamole-client/blob/fe6677bf4ebaa8662418013ab1af8c7060224ef5/guacamole/src/main/frontend/src/translations/zh.json

   "FIELD_OPTION_SERVER_LAYOUT_DE_CH_QWERTZ" : "Swiss German (Qwertz)",
   "FIELD_OPTION_SERVER_LAYOUT_DE_DE_QWERTZ" : "German (Qwertz)",
   "FIELD_OPTION_SERVER_LAYOUT_EMPTY"        : "",
   "FIELD_OPTION_SERVER_LAYOUT_EN_GB_QWERTY" : "UK English (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_EN_US_QWERTY" : "US English (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_ES_ES_QWERTY" : "Spanish (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_ES_LATAM_QWERTY" : "Latin American (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_FAILSAFE"     : "Unicode",
   "FIELD_OPTION_SERVER_LAYOUT_FR_BE_AZERTY" : "Belgian French (Azerty)",
   "FIELD_OPTION_SERVER_LAYOUT_FR_CA_QWERTY" : "Canadian French (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_FR_CH_QWERTZ" : "Swiss French (Qwertz)",
   "FIELD_OPTION_SERVER_LAYOUT_FR_FR_AZERTY" : "French (Azerty)",
   "FIELD_OPTION_SERVER_LAYOUT_HU_HU_QWERTZ" : "Hungarian (Qwertz)",
   "FIELD_OPTION_SERVER_LAYOUT_IT_IT_QWERTY" : "Italian (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_JA_JP_QWERTY" : "Japanese (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_NO_NO_QWERTY" : "Norwegian (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_PL_PL_QWERTY" : "Polish (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_PT_BR_QWERTY" : "Portuguese Brazilian (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_PT_PT_QWERTY" : "Portuguese (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_RO_RO_QWERTY" : "Romanian (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_SV_SE_QWERTY" : "Swedish (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_DA_DK_QWERTY" : "Danish (Qwerty)",
   "FIELD_OPTION_SERVER_LAYOUT_TR_TR_QWERTY" : "Turkish-Q (Qwerty)",
*/
