import * as express from "express";
import { Set } from "immutable";

import { privilegeDictionary } from "./priv-config";
import { getAuthentication } from "../../common";

interface IEndPointPrivilegeDescriptor {
    readonly [endPointREString: string]: {
        readonly [requestMethod: string]: string[]
    };
}

interface IEndPoint2REMap {
    readonly [kendPointREStringey: string]: RegExp;
}

type RequiredPrivilegesGetter = (url: string, requestMethod: string) => Set<string>;

const privilegesForEndPoints: IEndPointPrivilegeDescriptor = Object.freeze({
    "^/icon$": {
        POST: [
            privilegeDictionary.CREATE_ICON
        ]
    },
    // "/icon/:id/format/:format/size/:size"
    "^/icon/[^/]+/format/[^/]+/size/[^/]+$": {
        POST: [
            privilegeDictionary.CREATE_ICON,
            privilegeDictionary.ADD_ICON_FILE
        ]
    }
});

export const createPrivEndPointToREMap: (endPointPrivDesc: IEndPointPrivilegeDescriptor) => IEndPoint2REMap
= endPointPrivDesc => Object.keys(endPointPrivDesc).reduce(
    (acc, key) => Object.assign(acc, { [key]: new RegExp(key) }),
    {}
);

const privEndPoint2RE: {[key: string]: RegExp} = createPrivEndPointToREMap(privilegesForEndPoints);

export const requiredPrivilegesGetterProvider: (
    epPrivDesc: IEndPointPrivilegeDescriptor,
    ep2REMap: IEndPoint2REMap
) => RequiredPrivilegesGetter
= (epPrivDesc, ep2REMap) => (url, requestMethod) =>
    Set(Object.keys(epPrivDesc))
        .filter(route => ep2REMap[route].test(url))
        .flatMap(
            route => Set(Object.keys(epPrivDesc[route]))
                        .filter(privMethod => requestMethod === privMethod)
                        .flatMap(privMethod => Set(epPrivDesc[route][privMethod]))
        ).toSet();

const createDefaultRequiredPrivilegesGetter
    = () => requiredPrivilegesGetterProvider(privilegesForEndPoints, privEndPoint2RE);

const getRequiredPrivileges = createDefaultRequiredPrivilegesGetter();

const hasPrivilege: (endPointPrivileges: Set<string>, userPrivileges: Set<string>) => boolean
= (endPointPrivileges, userPrivileges) => endPointPrivileges.size === 0 ||
                                          endPointPrivileges.intersect(userPrivileges).size > 0;

export const hasRequiredPrivileges = (req: express.Request) => {
    const requiredPrivileges: Set<string> = getRequiredPrivileges(req.url, req.method);
    return requiredPrivileges.size === 0 ||
            getAuthentication(req.session) &&
                hasPrivilege(requiredPrivileges, getAuthentication(req.session).privileges);
};