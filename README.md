arguments and their applications:

cmd:
    - --verbose, -v: verbosity logging
    - --enterprise: create enterprise build using default options
    - --grafana: creates grafana build using default options
    - --enterprise-ref: create enterprise build using specified ref, also sets --enterprise=true
    - --enterprise-dir: create enterprise build using specified local dir, also sets --enterprise=true
    - --grafana-ref:  create grafana build using specified ref, also sets --grafana=true
    - --grafana-dir: create grafana build using specified local dir, also sets --grafana=true

local directory will take precedent over git hash


