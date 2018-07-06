import * as fs from "fs";
import * as path from "path";
import * as Process from "process";

import logger from "./utils/logger";
import * as Rx from "rxjs";
import { fileExists, readTextFile } from "./utils/rx";

const ICON_REPO_HOME = path.resolve(Process.env.HOME, ".ui-toolbox/icon-repo");

interface IServerConfiguration {
    readonly hostname: string;
    readonly port: number;
    readonly context?: string;
}

const configurationDataProto = Object.freeze({
    server_hostname: "",
    server_port: 0,
    server_url_context: "",
    app_description: "",
    path_to_static_files: "",
    icon_data_location_git: "",
    icon_data_allowed_formats: "",
    icon_data_allowed_sizes: "",
    authentication_type: "",
    oidc_client_id: "",
    oidc_client_secret: "",
    oidc_access_token_url: "",
    oidc_user_authorization_url: "",
    oidc_client_redirect_back_url: "",
    oidc_token_issuer: "",
    oidc_ip_jwt_public_key_url: "",
    oidc_ip_jwt_public_key_pem_base64: "",
    oidc_ip_logout_url: "",
    users_by_roles: {none: [""]},
    conn_host: "",
    conn_port: "",
    conn_user: "",
    conn_password: "",
    conn_database: "",
    enable_backdoors: false,
    logger_level: ""
});

export type ConfigurationData = typeof configurationDataProto;

const clone = (obj: any) => JSON.parse(JSON.stringify(obj));

const defaultSettings = {
    server_hostname: "localhost",
    server_port: 8090,
    server_url_context: "/",
    authentication_type: "oidc",
    app_description: "Collection of custom icons designed at Wombat Inc.",
    path_to_static_files: path.join(__dirname, "..", "..", "..", "client", "dist"),
    icon_data_location_git: path.resolve(ICON_REPO_HOME, "git-repo"),
    conn_host: "localhost",
    conn_port: "5432",
    conn_user: "iconrepo",
    conn_password: "iconrepo",
    conn_database: "iconrepo",
    icon_data_allowed_formats: "svg, png",
    icon_data_allowed_sizes: "18px, 24px, 48px, 18dp, 24dp, 36dp, 48dp, 144dp",
    enable_backdoors: false
};

export const getDefaultConfiguration: () => ConfigurationData = () => Object.assign(
    clone(configurationDataProto),
    clone(defaultSettings)
);

const ctxLogger = logger.createChild("appConfig");

export const DEFAULT_CONFIG_FILE_PATH = path.join(ICON_REPO_HOME, "config.json");

const getConfigFilePathByProfile: (configProfile: string) => string = configProfile => {
    return path.join(__dirname, "configurations", `${configProfile}.json`);
};

export const getConfigFilePath: () => string = () => {
    let result = null;
    if (process.env.ICON_REPO_CONFIG_FILE) {
        result = process.env.ICON_REPO_CONFIG_FILE;
    } else if (process.env.ICON_REPO_CONFIG_PROFILE) {
        result = getConfigFilePathByProfile(process.env.ICON_REPO_CONFIG_PROFILE);
    } else {
        result = DEFAULT_CONFIG_FILE_PATH;
    }
    ctxLogger.info("Configuration file: " + result);
    return result;
};

const configFilePath: string = getConfigFilePath();

const ignoreJSONSyntaxError: (error: any) => Rx.Observable<any> = error => {
    if (error instanceof SyntaxError) {
        ctxLogger.error("Skipping syntax error...");
        return Rx.Observable.of({});
    } else {
        throw error;
    }
};

export const updateConfigurationDataWithEnvVarValues = <T> (proto: T, conf: T) =>
    Object.keys(proto).reduce(
        (acc: any, key: string) => process.env[key.toUpperCase()]
            ? Object.assign(acc, {[key]: process.env[key.toUpperCase()]})
            : acc,
        conf
    );

export const readConfiguration: <T> (filePath: string, proto: T, defaults: any) => Rx.Observable<T>
= (filePath, proto, defaults) => {
    return fileExists(filePath)
        .flatMap(exists => {
            if (exists) {
                logger.info("Updating configuration from %s...", configFilePath);
                return readTextFile(filePath)
                    .map(fileContent => JSON.parse(fileContent))
                    .catch(error => ignoreJSONSyntaxError(error));
            } else {
                logger.warn("Configuration file doesn't exist: %s...", configFilePath);
                return Rx.Observable.of({});
            }
        })
        .map(json => Object.assign(clone(defaults), json))
        .do(conf => updateConfigurationDataWithEnvVarValues(proto, conf));
};

let configurationData: ConfigurationData;

export type ConfigurationDataProvider = () => ConfigurationData;

const updateState: () => Rx.Observable<ConfigurationDataProvider> = () => {
    return readConfiguration(configFilePath, configurationDataProto, defaultSettings)
        .do(conf => {
            if (conf.logger_level) {
                logger.setLevel(conf.logger_level);
            }
            configurationData = conf;
        })
        .map(conf => () => configurationData);
};

const watchConfigFile = (filePathToWatch: string) => {
    if (watcher != null) {
        watcher.close();
    }
    watcher = fs.watch(filePathToWatch, (event, filename) => {
        switch (event) {
            case "rename": // Editing with vim results in this event
                fileExists(filePathToWatch)
                .do(exists => {
                    logger.warn(
                        "Ooops! Configuration file was renamed?",
                        filePathToWatch
                    );
                })
                .filter(exists => exists)
                .do(b => updateState())
                .toPromise()
                    .then(
                        r => watchConfigFile(filePathToWatch),
                        err => ctxLogger.error("Configuration error", err)
                    );
                break;
            case "change":
                updateState();
                break;
        }
    });
};

Rx.Observable
    .forkJoin(Rx.Observable.of(configFilePath), fileExists(configFilePath))
    .filter(value => value[1])
        .do(value => watchConfigFile(value[0]))
        .subscribe();

let watcher: fs.FSWatcher = null;

export default updateState();
