package localcommand

func LinuxAuthenticatorTemplate() string {
	return `#!/bin/bash

# #### REQUEST INFO ####
# echo REQUEST INFO:
# json_pp

username="$1"
password="$2"

salt=$(getent shadow $username | cut -d$ -f3)
epassword=$(getent shadow $username | cut -d: -f2)

match=$(python -c 'import crypt; print crypt.crypt("'"$password"'", "$6$'$salt'")')

if [ "$match" == "$epassword" ]; then
  # success
  exit 0
fi

# failed
exit 1
`
}
