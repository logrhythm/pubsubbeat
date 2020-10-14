""" use siem repo 'workspace' file's go_repository info to generate 'go.mod' file
    Following rules on
        Confluence veracode page https://confluence.logrhythm.com/x/3tF4Aw
        Confluence gosec page https://confluence.logrhythm.com/x/Oed4Aw
"""

import sys
import re
import json
import getopt
from os.path import (abspath, dirname, realpath, isfile, basename, join)


class Require_Item:
    """
    go mod require item
    """

    def __init__(self):
        """ create object with default values """
        self.name = None     # go_repository's name
        self.package = None  # go mod replace package
        self.version = None  # go mod replace version
        self.comment = None  # in line comment
        self.extra = None    # extra comment line

    def set_all(self, r_name, r_package, r_version, r_comment, r_extra):
        """
        Set all Item values
        """
        self.name = r_name
        self.package = r_package
        self.version = r_version
        self.comment = r_comment
        self.extra = r_extra

    def copy(self):
        """
        return copy of item
        """
        new_item = Require_Item()
        new_item.set_all(self.name, self.package, self.version,
                         self.comment, self.extra)
        return new_item

    def __str__(self):
        self.to_string()

    def to_string(self, tab=False):
        """
        To Strng function
        Arguments
            tab {boolean} - when true start with tab; otherwise 4 space
        """
        start = "\t" if tab else "   "

        if self.package is None:
            return ""
        version = "" if self.version is None else self.version
        comment = "" if self.comment is None else self.comment
        rtn_string = start + self.package + " " + version + " // " + comment
        if self.extra is not None:
            rtn_string += "\n" + start + "// " + self.extra
        return rtn_string

    def set_via_ws(self, in_dict, over_json):
        """
        set item via on Workspace go_rep dictionary
            using override_json
        """
        self.set_all(None, None, None, None, None)
        if (in_dict.get('name') is None or
                in_dict.get('importpath') is None):
            return

        self.name = in_dict['name'].split('"')[1]

        self.package = in_dict.get('importpath').split('"')[1]

        if in_dict.get('version') is not None:
            self.version = in_dict['version'].split('"')[1]
            self.comment = "WS " + self.version

        elif in_dict.get('tag') is not None:
            self.version = in_dict['tag'].split('"')[1]
            self.comment = "WS " + self.version

        elif in_dict.get('commit') is not None:
            self.version = in_dict['commit'].split('"')[1]
            self.comment = "WS " + self.version

        elif in_dict.get('urls') is not None:
            path_string = in_dict['urls'].split('"')[1].split(":")[1]
            tar_gz_name = basename(path_string)
            self.version = tar_gz_name[0:-7]
            self.comment = "WS " + tar_gz_name

        else:
            pass

        if (over_json is not None and
                over_json.get(self.name) is not None):
            print("found override info: {}".format(self.package))
            self.version = over_json.get(self.name)["chg_verison"]
            self.comment += " " + over_json.get(self.name)["add_comment"]
            if over_json.get(self.name)["extra_cmt"] is not None:
                self.extra = over_json.get(self.name)["extra_cmt"]


def main(command, argv):
    """ return
        repo_name - via input -r or --repo argument
        workspace_file - real path to 'WORKSPACE' file
        override_file - real path to override go mod file based on repo
        override_name - override go mod file name
        out_file - real path to output go mod file based on repo
        silent - When True, not print go_repository information
    """
    VALID_NAME = "\t -r Valid repo_name: siem, pubsubbeat, sophoscentralbeat"
    SILENT_INFO = "\t -s To not print go_repository informationi (Default print)"
    HELP = '-r <repo_name> -s \n' + VALID_NAME + '\n' + SILENT_INFO
    main_repo_name = 'NONE'
    is_silent = False

    try:
        opts, argc = getopt.getopt(argv, "hr:s", ["repo=", "slient"])
        ignore_unused(argc)
    except getopt.GetoptError:
        print('Exit: {} {}'.format(command, HELP))
        sys.exit(1)
    for opt, arg in opts:
        if opt == '-h':
            print('Help: {} {}'.format(command, HELP))
            sys.exit(0)
        elif opt in ("-r", "--repo"):
            main_repo_name = arg
        elif opt in ("-s", "--slient"):
            is_silent = True
        else:
            print('Exit: {} {}'.format(command, HELP))
            sys.exit(3)

    if main_repo_name == 'NONE':
        print('Exit: {} {}'.format(command, HELP))
        sys.exit(4)

    # set default value
    rel_path_to_repo = join('..', '..', '..')
    base_dir_path = realpath(dirname(abspath(__file__)))
    repo_dir_path = realpath(join(base_dir_path, rel_path_to_repo))
    main_work_file = realpath(join(repo_dir_path, 'WORKSPACE'))
    main_over_file = None
    main_out_file = None

    if main_repo_name == 'siem':
        rel_cmd_path = '..'
        main_over_name = "oc_beats_override.json"
        main_over_file = realpath(join(base_dir_path, rel_cmd_path, main_over_name))
        main_out_file = realpath(join(base_dir_path, rel_cmd_path, "oc_beats_go.mod"))
    elif main_repo_name == 'pubsubbeat':
        rel_cmd_path = join('..', 'pubsubbeat')
        main_over_name = "pubsub_override.json"
        main_over_file = realpath(join(base_dir_path, rel_cmd_path, main_over_name))
        main_out_file = realpath(join(base_dir_path, rel_cmd_path, "pubsub_go.mod"))
    elif main_repo_name == 'sophoscentralbeat':
        rel_cmd_path = join('..', 'sophoscentralbeat')
        main_over_name = "sophos_override.json"
        main_over_file = realpath(join(base_dir_path, rel_cmd_path, main_over_name))
        main_out_file = realpath(join(base_dir_path, rel_cmd_path, "sophos_go.mod"))
    else:
        print("Exit: not valid repo {}".format(main_repo_name))
        print(VALID_NAME)
        sys.exit(5)
    return (main_repo_name, main_work_file, main_over_file,
            main_over_name, main_out_file, is_silent)


def ignore_unused(var):
    """ avoids pylint unused variable """
    tmp_a = var
    var = tmp_a


def set_mod_header(module_name, over_name):
    """
    set go.mod header lines
    """
    module_string = "module " + module_name
    blank_line = ""
    go_version = "go 1.12"
    created_info = r"// Automatic generated via repo: 'siem', file: 'WORKSPACE',"
    created_info += " override: '" + over_name + "'"
    replace_base = "replace " + r"github.com/logrhythm/" + module_name
    top_level_comment = r"// Assume at repo top level (" + module_name
    top_level_comment += r" directory) for 'go mod vendor' command"
    header_array = [module_string,
                    blank_line,
                    go_version,
                    blank_line,
                    created_info,
                    blank_line,
                    top_level_comment]

    if module_name == 'sophoscentralbeat':
        other_replace = replace_base + r"/sophoscentral latest => ./sophoscentral"
        header_array.append(other_replace)

    replace_line = replace_base + r" latest => ./"
    header_array.append(replace_line)
    header_array.append(blank_line)
    return header_array


def init_require_list(module_name):
    """ return initial go mod require list """
    out_list = []
    req_item = Require_Item()
    if module_name == 'sophoscentralbeat':
        req_item.set_all(
            "", r"github.com/logrhythm/" + module_name + r"/sophoscentral",
            "latest", "WS use above replace path", None)
        out_list.append(req_item.copy())

    req_item.set_all("", r"github.com/logrhythm/" + module_name,
                     "latest", "WS use above replace path", None)
    out_list.append(req_item.copy())

    req_item.set_all("", r"github.com/Sirupsen/logrus",
                     "1.0.2", "required not in WS", None)
    out_list.append(req_item.copy())

    req_item.set_all("", r"github.com/golang/protobuf",
                     "v1.3.0 ", "required not in WS", None)
    out_list.append(req_item.copy())

    req_item.set_all("", None, None, None, None)
    out_list.append(req_item.copy())

    if module_name != 'siem':
        req_item.set_all("", "", "",
                         "similar to siem cmd/veracode/veracode_ws_go.mod file",
                         None)
        out_list.append(req_item.copy())
    return out_list


if __name__ == "__main__":
    (repo_name, workspace_file,
     override_file, override_short,
     output_file, silent) = main(sys.argv[0], sys.argv[1:])

    if not isfile(workspace_file):
        print("Exit not found WORKSPACE: {}".format(workspace_file))
        sys.exit(2)

    override_json = None
    if isfile(override_file):
        with open(override_file) as f:
            override_json = json.load(f)

    go_repository_start = r'^go_repository\('
    go_repository_end = r'^\)'
    print("repo name: {}".format(repo_name))
    print("workspace: {}".format(basename(workspace_file)))
    print("override:  {}".format(basename(override_file)))

    in_file = open(workspace_file, mode='r', encoding='utf-8')
    lines = in_file.readlines()
    in_file.close()

    go_repository_count = 0
    go_repository_list = []
    go_repository_dict = {}
    in_go_repository = False
    newline = "\n"

    go_mod_header = set_mod_header(repo_name, override_short)

    go_mod_require_list = init_require_list(repo_name)

    out_item = Require_Item()
    for line in lines:
        if in_go_repository:
            if re.match(go_repository_end, line):
                in_go_repository = False
                for k, v in go_repository_dict.items():
                    if not silent:
                        print("   key: {}; value: {}".format(k, v))
                out_item.set_via_ws(go_repository_dict, override_json)
                if not silent:
                    print(out_item.to_string())
                go_repository_list.append(go_repository_dict)
                go_mod_require_list.append(out_item.copy())
                if not silent:
                    print("done  {}:".format(go_repository_count))
            else:
                parts = line.split('=')
                parts = [p.strip() for p in parts]
                go_repository_dict[parts[0]] = parts[1]
        else:
            if re.match(go_repository_start, line):
                in_go_repository = True
                go_repository_dict = {}
                go_repository_count += 1
                if not silent:
                    print("start {}: {}".format(go_repository_count, go_repository_start))

    # write go_mod_dict to file
    out_fp = open(output_file, mode='w', encoding='utf-8')
    print("go mod:    {}".format(basename(output_file)))

    for header in go_mod_header:
        out_fp.write(header + newline)

    replace_list_start = r"require ("
    replace_list_end = r")"
    out_fp.write(replace_list_start + newline)
    for go_mod_item in go_mod_require_list:
        out_fp.write(go_mod_item.to_string(tab=True) + newline)
    out_fp.write(replace_list_end + newline)
    print("")
