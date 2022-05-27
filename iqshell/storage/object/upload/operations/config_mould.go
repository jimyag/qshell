package operations

const uploadConfigMouldJsonString = `{
	"log_level": "debug",
	"log_file": "",
	"log_rotate": 10,
	"log_stdout": "true",
	"up_host": "",
	"src_dir": "<select one between src_dir and file_list>",
	"file_list": "<select one between src_dir and file_list>",
	"ignore_dir": "",
	"skip_file_prefixes": "",
	"skip_path_prefixes": ".,..",
	"skip_fixed_strings": "",
	"skip_suffixes": "",
	"file_encoding": "",
	"bucket": "",
	"resumable_api_v2": "false",
	"resumable_api_v2_part_size": 1048576,
	"put_threshold": 16777216,
	"key_prefix": "",
	"overwrite": "true",
	"check_exists": "true",
	"check_hash": "false",
	"check_size": "true",
	"rescan_local": "false",
	"file_type": 0,
	"delete_on_success": "false",
	"disable_resume": "false",
	"disable_form": "false",
	"record_root": "",
}`