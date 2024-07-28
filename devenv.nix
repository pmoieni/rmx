{
  pkgs,
  lib,
  # config,
  # inputs,
  ...
}: {
  # env
  env.GREET = "RMX devenv";

  # packages
  packages = [pkgs.git];

  # scripts
  scripts.hello.exec = "echo $GREET";

  # startup
  enterShell = ''
    hello
  '';

  # tests
  enterTest = ''
    echo "Running tests"
    git --version
  '';

  # services
  services.postgres = {
    enable = true;
    initialScript = ''
      CREATE ROLE postgres SUPERUSER;
      CREATE ROLE "rmx" WITH LOGIN PASSWORD "rmx";
    '';
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

  # languages
  languages.nix.enable = true;
  languages.go.enable = true;

  # pre-commit hooks
  # pre-commit.hooks.shellcheck.enable = true;

  # processes
  # processes.ping.exec = "ping example.com";
}
