package paths

const (
	SERVER_DIR       = "/etc/wireguard/"
	WG_MANAGER_DIR   = SERVER_DIR + ".wg_manager"
	USERS_CONFIG_DIR = SERVER_DIR + ".configs"
	USERS_DIR        = SERVER_DIR + "users"
	TC_DIR           = SERVER_DIR + ".tc"
	SSP_DIR          = SERVER_DIR + ".ssp"
	TC_CLASS_FILE    = "classes"
	TC_FILTER_FILE   = "filters"
	TC_FILE          = "tc"
	TC_CONFIG_FILE   = "tc.sh"
	TC_SERVICE_FILE  = "tc.service"
)
