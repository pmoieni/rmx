{
  pkgs,
  lib,
  config,
  inputs,
  ...
}:
{
  env.GREET = "RMX devenv";

  # packages
  packages = [ pkgs.git ];

  # git-hooks
  git-hooks.hooks = {
    shellcheck.enable = true;
    treefmt.enable = true;
  };

  # languages
  languages.nix.enable = true;
  languages.go.enable = true;

  # processes

  # services
  services.postgres = {
    enable = true;
    initialDatabases = [
      {
        name = "rmx-dev";
        user = "rmx-dev";
        pass = "rmx-dev";
      }
    ];
    listen_addresses = "127.0.0.1";
    port = 5432;
    settings = {
      log_connections = true;
      log_statement = "all";
      logging_collector = true;
      log_disconnections = true;
      log_destination = lib.mkForce "syslog";
    };
  };

  # https://devenv.sh/tasks/
  # tasks = {
  #   "myproj:setup".exec = "mytool build";
  #   "devenv:enterShell".after = [ "myproj:setup" ];
  # };

  # temp fix
  cachix.enable = false;
}
