package analysis

import (
	"log"
	"strings"
)

func SplitStringWithPosition(input string) map[string][]int {
	words := make(map[string][]int)
	currentWord := ""
	startPosition := -1

	formattedInp := strings.ReplaceAll(input, "(", " ")
	formattedInp = strings.ReplaceAll(formattedInp, ")", " ")

	for i, char := range formattedInp {
		if char != ' ' {
			if startPosition == -1 {
				startPosition = i
			}
			currentWord += string(char)
		} else if startPosition != -1 {
			words[currentWord] = append(words[currentWord], startPosition)
			currentWord = ""
			startPosition = -1
		}
	}

	// Check for the last word
	if startPosition != -1 {
		words[currentWord] = append(words[currentWord], startPosition)
	}

	return words
}

func getOpenCloseQuotePairs(input, searchChar string, logger *log.Logger) ([]int, []int) {
	var startIndex []int
	var endIndex []int

	lookingForOpen := true
	for i, c := range input {
		if c == rune(searchChar[0]) {
			if lookingForOpen {
				lookingForOpen = false

				startIndex = append(startIndex, i)
			} else {
				lookingForOpen = true
				endIndex = append(endIndex, i)
			}

		}

	}
	return startIndex, endIndex
}

func CreateStringTokens(input string, startLineNo int, searchChar string, logger *log.Logger) []token {

	startIndex, endIndex := getOpenCloseQuotePairs(input, searchChar, logger)

	var tokenList []token

	if len(startIndex) == 0 || len(endIndex) == 0 {
		return tokenList
	}

	defaultType := 3
	defaultModifiers := 0

	for i := range endIndex {

		stringInput := input[startIndex[i]:endIndex[i]]
		stringLines := strings.Split(input[:startIndex[i]], "\n")
		stringLinesTruncated := stringLines[:len(stringLines)-1]
		linesBefore := startLineNo + len(stringLinesTruncated) - 2

		adjustedStartIndex := 0

		for _, l := range strings.Split(input, "\n") {
			if strings.Contains(l, strings.Split(stringInput, "\n")[0]) {
				adjustedStartIndex = strings.Index(l, strings.Split(stringInput, "\n")[0])
			}
		}

		linesIn := strings.Split(stringInput, "\n")

		if len(linesIn) == 1 {
			tokenList = append(tokenList, token{
				tokenValue:     stringInput,
				absStartIndex:  adjustedStartIndex,
				absLineNo:      linesBefore + 1,
				length:         len(stringInput) + 1,
				tokenType:      &defaultType,
				tokenModifiers: &defaultModifiers,
			})
		} else {
			for j, l := range linesIn {
				if j != 0 {

					tokenList = append(tokenList, token{
						tokenValue:     l,
						absStartIndex:  7,
						absLineNo:      linesBefore + 1 + j,
						length:         len(l) + 1 - 7,
						tokenType:      &defaultType,
						tokenModifiers: &defaultModifiers,
					})
				} else {

					tokenList = append(tokenList, token{
						tokenValue:     l,
						absStartIndex:  adjustedStartIndex,
						absLineNo:      linesBefore + 1,
						length:         len(l) + 1 + 7,
						tokenType:      &defaultType,
						tokenModifiers: &defaultModifiers,
					})
				}

			}

		}
	}
	return tokenList

}

func GetSqlTokens() []string {
	return []string{
		"alter catalog",
		"alter",
		"alter connection",
		"alter credential",
		"alter database",
		"alter location",
		"alter provider",
		"alter recipient",
		"alter streaming table",
		"alter table",
		"alter schema",
		"alter share",
		"alter view",
		"alter volume",
		"comment on",
		"create bloomfilter index",
		"create",
		"create catalog",
		"create connection",
		"create database",
		"create function (sql)",
		"create function (external)",
		"create location",
		"create materialized view",
		"create recipient",
		"create schema",
		"create server",
		"create share",
		"create streaming table",
		"create table",
		"create view",
		"create volume",
		"declare variable",
		"drop bloomfilter index",
		"drop catalog",
		"drop connection",
		"drop database",
		"drop credential",
		"drop function",
		"drop location",
		"drop provider",
		"drop recipient",
		"drop schema",
		"drop share",
		"drop table",
		"drop",
		"table",
		"view",
		"schema",
		"catalog",
		"drop variable",
		"drop view",
		"drop volume",
		"msck repair table",
		"refresh foreign (catalog, schema, or table)",
		"refresh (materialized view or streaming table)",
		"sync",
		"truncate table",
		"undrop table",
		"copy into",
		"delete from",
		"truncate",
		"merge",
		"with",
		"as",
		"into",
		"from",
		"insert into",
		"insert",
		"into",
		"overwrite",
		"insert overwrite directory",
		"insert overwrite directory with hive format",
		"load data",
		"merge into",
		"update",
		"query",
		"select",
		"values",
		"explain",
		"cache select",
		"convert to delta",
		"describe history",
		"fsck repair table",
		"generate",
		"optimize",
		"reorg table",
		"restore",
		"vacuum",
		"analyze table",
		"cache table",
		"clear cache",
		"refresh cache",
		"refresh function",
		"refresh table",
		"uncache table",
		"describe catalog",
		"describe",
		"describe connection",
		"describe credential",
		"describe database",
		"describe function",
		"describe location",
		"describe provider",
		"describe query",
		"describe recipient",
		"describe schema",
		"describe share",
		"describe table",
		"describe volume",
		"list",
		"show all in share",
		"show catalogs",
		"show columns",
		"show connections",
		"show create table",
		"show credentials",
		"show databases",
		"show functions",
		"show groups",
		"show locations",
		"show partitions",
		"show providers",
		"show recipients",
		"show schemas",
		"show shares",
		"show shares in provider",
		"show table",
		"show tables",
		"show tables dropped",
		"show tblproperties",
		"show users",
		"show views",
		"show volumes",
		"execute immediate",
		"reset",
		"set",
		"set timezone",
		"set variable",
		"use catalog",
		"use database",
		"use schema",
		"add archive",
		"add file",
		"add jar",
		"list archive",
		"list file",
		"list jar",
		"alter group",
		"create group",
		"deny",
		"drop group",
		"grant",
		"grant share",
		"repair privileges",
		"revoke",
		"revoke share",
		"show grants",
		"show grants on share",
		"show grants to recipient",
	}
}

func GetSqlFunctions() []string {
	return []string{"abs",
		"acos",
		"acosh",
		"add_months",
		"aes_decrypt",
		"aes_encrypt",
		"aggregate",
		"ai_analyze_sentiment",
		"ai_classify",
		"ai_extract",
		"ai_fix_grammar",
		"ai_gen",
		"ai_generate_text",
		"ai_mask",
		"ai_query",
		"ai_similarity",
		"ai_summarize",
		"ai_translate",
		"&",
		"and",
		"any",
		"any_value",
		"approx_count_distinct",
		"approx_percentile",
		"approx_top_k",
		"array",
		"array_agg",
		"array_append",
		"array_compact",
		"array_contains",
		"array_distinct",
		"array_except",
		"array_insert",
		"array_intersect",
		"array_join",
		"array_max",
		"array_min",
		"array_position",
		"array_prepend",
		"array_remove",
		"array_repeat",
		"array_size",
		"array_sort",
		"array_union",
		"arrays_overlap",
		"arrays_zip",
		"ascii",
		"asin",
		"asinh",
		"assert_true",
		"*",
		"atan",
		"atan2",
		"atanh",
		"avg",
		"!=",
		"!",
		"base64",
		"between",
		"bigint",
		"bin",
		"binary",
		"bit_and",
		"bit_count",
		"bit_get",
		"bit_length",
		"bit_or",
		"bit_reverse",
		"bit_xor",
		"bitmap_bit_position",
		"bitmap_bucket_number",
		"bitmap_construct_agg",
		"bitmap_count",
		"bitmap_or_agg",
		"bool_and",
		"bool_or",
		"boolean",
		"[",
		"bround",
		"btrim",
		"cardinality",
		"^",
		"case",
		"cast",
		"cbrt",
		"ceil",
		"ceiling",
		"char",
		"char_length",
		"character_length",
		"charindex",
		"chr",
		"cloud_files_state",
		"coalesce",
		"collect_list",
		"collect_set",
		"::",
		":",
		"concat",
		"concat_ws",
		"contains",
		"conv",
		"convert_timezone",
		"corr",
		"cos",
		"cosh",
		"cot",
		"count",
		"count_if",
		"count_min_sketch",
		"covar_pop",
		"covar_samp",
		"crc32",
		"csc",
		"cube",
		"cume_dist",
		"curdate",
		"current_catalog",
		"current_database",
		"current_date",
		"current_metastore",
		"current_recipient",
		"current_schema",
		"current_timestamp",
		"current_timezone",
		"current_user",
		"current_version",
		"date",
		"date_add",
		"date_add",
		"date_diff",
		"date_format",
		"date_from_unix_date",
		"date_part",
		"date_sub",
		"date_trunc",
		"dateadd",
		"dateadd",
		"datediff",
		"datediff",
		"day",
		"dayofmonth",
		"dayofweek",
		"dayofyear",
		"decimal",
		"decode",
		"decode",
		"degrees",
		"dense_rank",
		"div",
		".",
		"double",
		"e",
		"element_at",
		"elt",
		"encode",
		"endswith",
		"==",
		"=",
		"equal_null",
		"event_log",
		"every",
		"exists",
		"exp",
		"explode",
		"explode_outer",
		"expm1",
		"extract",
		"factorial",
		"filter",
		"find_in_set",
		"first",
		"first_value",
		"flatten",
		"float",
		"floor",
		"forall",
		"format_number",
		"format_string",
		"from_csv",
		"from_json",
		"from_unixtime",
		"from_utc_timestamp",
		"from_xml",
		"get",
		"get_json_object",
		"getbit",
		"getdate",
		"greatest",
		"grouping",
		"grouping_id",
		">=",
		">",
		"h3_boundaryasgeojson",
		"h3_boundaryaswkb",
		"h3_boundaryaswkt",
		"h3_centerasgeojson",
		"h3_centeraswkb",
		"h3_centeraswkt",
		"h3_compact",
		"h3_coverash3",
		"h3_coverash3string",
		"h3_distance",
		"h3_h3tostring",
		"h3_hexring",
		"h3_ischildof",
		"h3_ispentagon",
		"h3_isvalid",
		"h3_kring",
		"h3_kringdistances",
		"h3_longlatash3",
		"h3_longlatash3string",
		"h3_maxchild",
		"h3_minchild",
		"h3_pointash3",
		"h3_pointash3string",
		"h3_polyfillash3",
		"h3_polyfillash3string",
		"h3_resolution",
		"h3_stringtoh3",
		"h3_tessellateaswkb",
		"h3_tochildren",
		"h3_toparent",
		"h3_try_distance",
		"h3_try_polyfillash3",
		"h3_try_polyfillash3string",
		"h3_try_validate",
		"h3_uncompact",
		"h3_validate",
		"hash",
		"hex",
		"hll_sketch_agg",
		"hll_sketch_estimate",
		"hll_union",
		"hll_union_agg",
		"hour",
		"hypot",
		"if",
		"iff",
		"ifnull",
		"ilike",
		"in",
		"initcap",
		"inline",
		"inline_outer",
		"input_file_block_length",
		"input_file_block_start",
		"input_file_name",
		"instr",
		"int",
		"is_account_group_member",
		"is_member",
		"is",
		"is",
		"isnan",
		"isnotnull",
		"isnull",
		"is",
		"is",
		"java_method",
		"json_array_length",
		"json_object_keys",
		"json_tuple",
		"kurtosis",
		"lag",
		"last",
		"last_day",
		"last_value",
		"lcase",
		"lead",
		"least",
		"left",
		"len",
		"length",
		"levenshtein",
		"like",
		"list_secrets",
		"ln",
		"locate",
		"log",
		"log10",
		"log1p",
		"log2",
		"lower",
		"lpad",
		"<=>",
		"<=",
		"<>",
		"ltrim",
		"<",
		"luhn_check",
		"make_date",
		"make_dt_interval",
		"make_interval",
		"make_timestamp",
		"make_ym_interval",
		"map",
		"map_concat",
		"map_contains_key",
		"map_entries",
		"map_filter",
		"map_from_arrays",
		"map_from_entries",
		"map_keys",
		"map_values",
		"map_zip_with",
		"mask",
		"max",
		"max_by",
		"md5",
		"mean",
		"median",
		"min",
		"min_by",
		"-",
		"-",
		"minute",
		"mod",
		"mode",
		"monotonically_increasing_id",
		"month",
		"months_between",
		"named_struct",
		"nanvl",
		"negative",
		"next_day",
		"not",
		"now",
		"nth_value",
		"ntile",
		"nullif",
		"nvl",
		"nvl2",
		"octet_length",
		"or",
		"overlay",
		"parse_url",
		"percent_rank",
		"percentile",
		"percentile_approx",
		"percentile_cont",
		"percentile_disc",
		"%",
		"pi",
		"||",
		"|",
		"+",
		"+",
		"pmod",
		"posexplode",
		"posexplode_outer",
		"position",
		"positive",
		"pow",
		"power",
		"printf",
		"quarter",
		"radians",
		"raise_error",
		"rand",
		"randn",
		"random",
		"range",
		"rank",
		"read_files",
		"read_kafka",
		"read_kinesis",
		"read_pubsub",
		"read_pulsar",
		"read_state_metadata",
		"read_statestore",
		"reduce",
		"reflect",
		"regexp",
		"regexp_count",
		"regexp_extract",
		"regexp_extract_all",
		"regexp_instr",
		"regexp_like",
		"regexp_replace",
		"regexp_substr",
		"regr_avgx",
		"regr_avgy",
		"regr_count",
		"regr_intercept",
		"regr_r2",
		"regr_slope",
		"regr_sxx",
		"regr_sxy",
		"regr_syy",
		"repeat",
		"replace",
		"reverse",
		"right",
		"rint",
		"rlike",
		"round",
		"row_number",
		"rpad",
		"rtrim",
		"schema_of_csv",
		"schema_of_json",
		"schema_of_json_agg",
		"schema_of_xml",
		"sec",
		"second",
		"secret",
		"sentences",
		"sequence",
		"session_user",
		"session_window",
		"sha",
		"sha1",
		"sha2",
		"shiftleft",
		"shiftright",
		"shiftrightunsigned",
		"shuffle",
		"sign",
		"signum",
		"sin",
		"sinh",
		"size",
		"skewness",
		"/",
		"slice",
		"smallint",
		"some",
		"sort_array",
		"soundex",
		"space",
		"spark_partition_id",
		"split",
		"split_part",
		"sql_keywords",
		"sqrt",
		"stack",
		"startswith",
		"std",
		"stddev",
		"stddev_pop",
		"stddev_samp",
		"str_to_map",
		"string",
		"struct",
		"substr",
		"substring",
		"substring_index",
		"sum",
		"table_changes",
		"tan",
		"tanh",
		"~",
		"timediff",
		"timestamp",
		"timestamp_micros",
		"timestamp_millis",
		"timestamp_seconds",
		"timestampadd",
		"timestampdiff",
		"tinyint",
		"to_binary",
		"to_char",
		"to_csv",
		"to_date",
		"to_json",
		"to_number",
		"to_timestamp",
		"to_unix_timestamp",
		"to_utc_timestamp",
		"to_varchar",
		"to_xml",
		"transform",
		"transform_keys",
		"transform_values",
		"translate",
		"trim",
		"trunc",
		"try_add",
		"try_aes_decrypt",
		"try_avg",
		"try_cast",
		"try_divide",
		"try_element_at",
		"try_multiply",
		"try_reflect",
		"try_subtract",
		"try_sum",
		"try_to_binary",
		"try_to_number",
		"try_to_timestamp",
		"typeof",
		"ucase",
		"unbase64",
		"unhex",
		"unix_date",
		"unix_micros",
		"unix_millis",
		"unix_seconds",
		"unix_timestamp",
		"upper",
		"url_decode",
		"url_encode",
		"user",
		"uuid",
		"var_pop",
		"var_samp",
		"variance",
		"version",
		"weekday",
		"weekofyear",
		"width_bucket",
		"window",
		"window_time",
		"xpath",
		"xpath_boolean",
		"xpath_double",
		"xpath_float",
		"xpath_int",
		"xpath_long",
		"xpath_number",
		"xpath_short",
		"xpath_string",
		"xxhash64",
		"year",
		"zip_with"}
}
