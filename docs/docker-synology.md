# Synology Docker Instructions

If you want to run coredns-omada on a Synology NAS you have the option of launching the Docker container from the terminal (SSH) with the standard Docker run command. \
However, there are several advantages to creating a container via the Docker plugin in DSM and launching from there. \
To do that, we need fit all the components of this Docker run command into the relevant fields in the DSM Docker plugin:

```
docker run \
--rm -it -m 128m \
--expose=53 --expose=53/udp -p 53:53 -p 53:53/udp \
--env OMADA_URL="<OMADA_URL>" \
--env OMADA_SITE="<OMADA_SITE>" \
--env OMADA_USERNAME="<OMADA_USERNAME>" \
--env OMADA_PASSWORD="<OMADA_PASSWORD>" \
--env OMADA_DISABLE_HTTPS_VERIFICATION="false" \
--env UPSTREAM_DNS="8.8.8.8" \
ghcr.io/dougbw/coredns_omada:latest
```

1. Install `Docker` from the Synology Package Center

2. You need to pull the container image onto the Synology. Use this guide as a reference: https://www.tobyscott.dev/blog/ghcr-synology-container-manager/

* `ssh <user>@synologynas`
* `sudo docker pull ghcr.io/dougbw/coredns_omada:latest`


3. Double click the coredns-omada Image to create a Docker Container
    - Select the `bridge` network from the Network pane, click Next.
    - <img src="./images/synology/network_pane.png" alt=“” width="40%" height="40%">
    - On the General Settings pane
        - Set the Container Name to whatever you want, default value is fine.
        - Enable auto restart
        - You can set system resource limitations if you wish
        - <img src="./images/synology/general_settings_pane.png" alt=“” width="40%" height="40%">
        - Click the Advanced Settings button
            - On the Environment tab, add an environment variable corresponding to each of the `--env` flags in the `docker run` command above. 
                - Environment variable values should not be surrounded by quotes. 
                - The IP address for the controller should be preceded by `https://`
            - Your Environment Variables screen should look like this once complete:
            - <img src="./images/synology/advanced_settings_env_vars.png" alt=“” width="40%" height="40%">
            - Save the Advanced Settings and click next on the General Settings pane.
            
	   
    - On the Port Settings pane you must change the Local ports from Auto to 53 for both TCP and UDP.
    - <img src="./images/synology/port_settings_pane.png" alt=“” width="40%" height="40%">
    - On the next pane, Volume Settings, click "Add File" and select the `Corefile` you created under the docker share above. Set the Mount Path to `/etc/coredns/Corefile`
    - Save everything and launch the newly created Container.
