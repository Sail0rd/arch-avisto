module main

import brawdunoir.prompts
import term
import os
import net.http
import time

// Command to upgrade all the OS and packages. It should be run before any other `paru` commands
const upgrade_command_string = 'sudo -u ${old_username} paru -Syu --skipreview --noconfirm'
// Username of the preconfigured user
const old_username = 'arch'
// Username of the login user
const login_username = 'login'
// Profiles available to the user with a preset of packages. Adding a profile here requires to add a match entry in the script.
const profiles = ['Dev', 'DevOps']
// Common and profile specific packages list.
const common_packages = ['duf', 'dust', 'helm', 'helmfile', 'jq', 'k9s', 'kubectl', 'micro', 'navi',
	'pre-commit', 'skaffold', 'unzip', 'wget']
const dev_packages = ['gitleaks']
const devops_packages = ['ansible', 'bottom', 'htop', 'iperf', 'gnu-netcat', 'net-tools', 'pgcli',
	'screen', 'sshuttle', 'tcpdump', 'inetutils', 'terraform', 'tmux']
// Environment variable with version in it
const version_env_key = 'ARCHAVISTO_VERSION'
// Path to the file, if it exist, that will prevent this script to run
const skip_file = '/opt/skip_bootscript'
// Connectivity test address (without https)
const test_address = 'google.com'
// VPN Toolkit systemd service name
const vpntoolkit_systemd_name = 'wsl-vpnkit.service'
// Path to wsl.conf file
const wsl_conf_file = '/etc/wsl.conf'

fn main() {
	if os.is_file(skip_file) {
		println('Skipping execution. Remove ${skip_file} file to execute again.')
		exit(0)
	}

	term.clear()
	version := os.getenv_opt(version_env_key)

	println('Welcome to your ${term.bright_cyan('Avisto Arch Linux')} WSL.')
	if version != none {
		println('Version: ${version}')
	}

	// Check connectivity, because of the VPN, the daemon vpn-toolkit should start in order to the image to have access to internet
	println('Checking for network connectivity (can fail on first startup due to VPN)')
	for i in 1 .. 11 {
		response := http.get('https://${test_address}') or { http.Response{} }
		if i == 10 {
			panic(term.fail_message('Unable to join Internet. Contact a DevOps internal member by Teams or by email devops-support@advans-group.atlassian.net'))
		} else if response == http.Response{} || response.status().is_error() {
			println(term.warn_message('Cannot GET ${test_address}, retrying in 5 seconds (attempt ${i}/10)'))
			time.sleep(5 * time.second)
		} else {
			println(term.ok_message('Network OK'))
			break
		}
		// Try to restart service
		if i == 2 {
			os.execute('systemctl reload-or-restart ${vpntoolkit_systemd_name}')
		}
	}

	// Upgrade OS
	background_update_prompt := prompts.confirm('Update the packages in the background?',
		true)
	if !background_update_prompt {
		println(term.fail_message('Cannot continue without updating the packages'))
		exit(1)
	}
	background_update_thread := spawn os.execute(upgrade_command_string)

	mut new_username := prompts.input('How should I call you?').to_lower()
	if new_username == '' {
		new_username = old_username
	}

	// Get profile and set the according packages to install additionnally.
	profile := prompts.choice('What profile do you want?', profiles)
	mut packages_to_prompt := common_packages.clone()
	match profile {
		'Dev' { packages_to_prompt << dev_packages }
		'DevOps' { packages_to_prompt << devops_packages }
		else { panic(term.fail_message('Profile is unknown')) }
	}
	packages_to_install := prompts.multichoice('Select packages you want to install',
		packages_to_prompt)

	// Wait for upgrade command launched at the beginning
	println('Waiting for your packages to be up to date (could be long, please wait)')
	background_update_result := background_update_thread.wait()

	if background_update_result.exit_code != 0 {
		panic(term.fail_message(background_update_result.output))
	}
	println(term.ok_message('System up to date!'))

	if packages_to_install.len > 0 {
		// Install profile dependant packages
		println("Installing the packages you've selected previously")
		packages_execute_result := os.execute(paru_install_command(packages_to_install))
		if packages_execute_result.exit_code != 0 {
			panic(term.fail_message(packages_execute_result.output))
		}
	}

	// Change username
	change_username_result := os.execute(change_username_command_string(old_username,
		new_username))
	if change_username_result.exit_code != 0 {
		panic(term.fail_message(change_username_result.output))
	}

	// // changing wsl is not working at the moment
	// wsl_conf_content := os.read_file(wsl_conf_file) or { panic(term.fail_message('Cannot read file ${wsl_conf_file}')) }
	// wsl_conf_updated_content := wsl_conf_content.replace_once(login_username, new_username)
	// os.write_file(wsl_conf_file, wsl_conf_updated_content) or { panic(term.fail_message('Cannot write into file ${wsl_conf_file}'))}

	// End messages
	println(term.ok_message('Installation finished!'))
	println('You can now go back to ${term.blue('Powershell')} and start WSL using ' +
		term.magenta('wsl -u ${new_username} -d <distro-name>'))

	// Assure we do not execute this another time
	mut file := os.create(skip_file) or { panic(term.fail_message('Cannot create ${skip_file}')) }
	file.close()
}

fn change_username_command_string(old string, new string) string {
	return 'sudo usermod --login=${new} --move-home --home=/home/${new} ${old}'
}

fn paru_install_command(packages []string) string {
	return 'sudo -u ${old_username} paru -S --noconfirm ${packages.join(' ')}'
}
