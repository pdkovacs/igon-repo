import getEndpointUrl from "./url";
import { throwError } from "./errors";
import { Set, List } from "immutable";

export interface IconfileDescriptor {
    readonly format: string;
    readonly size: string;
}

export interface IngestedIconfileDTO extends IconfileDescriptor {
    iconName: string;
    path: string;
}

export interface IconPath extends IconfileDescriptor {
    readonly path: string;
}

export interface IconPathWithUrl extends IconPath {
    readonly url: string;
}

export interface IconDescriptor {
    readonly name: string;
    readonly modifiedBy: string;
    readonly paths: Set<IconPath>;
    readonly tags: Set<string>;
}

export interface IconDTO {
    readonly name: string;
    readonly modifiedBy: string;
    readonly paths: IconPath[];
    readonly tags: string[];
}

export const describeAllIcons: () => Promise<Set<IconDescriptor>>
= () => fetch(getEndpointUrl(`/icon`), {
    method: "GET",
    credentials: "include"
})
.then(response => {
    if (response.status !== 200) {
        return throwError("Failed to fetch icon descriptions", response);
    } else {
        return response.json();
    }})
.then((iconDtoArray: IconDTO[]) => Set(iconDtoArray.map(iconDto => ({
        name: iconDto.name,
        modifiedBy: iconDto.modifiedBy,
        paths: Set(iconDto.paths),
        tags: Set(iconDto.tags)
    }))));

export const describeIcon = (iconName: string) => fetch(getEndpointUrl(`/icon/${iconName}`), {
    method: "GET",
    credentials: "include"
})
.then(response => {
    if (response.status !== 200) {
        return throwError("Failed to fetch icon description", response);
    } else {
        return response.json();
    }
});

export const createIcon: (formData: FormData) => Promise<IconDTO>
= formData => fetch(getEndpointUrl("/icon"), {
    method: "POST",
    credentials: "include",
    body: formData
})
.then(response => {
    if (response.status !== 201) {
        return throwError("Failed to create icon", response);
    } else {
        return response.json();
    }
});

export const ingestIconfile: (iconName: string, formData: FormData) => Promise<IngestedIconfileDTO>
= (iconName, formData) => fetch(getEndpointUrl(`/icon/${iconName}`), {
    method: "POST",
    credentials: "include",
    body: formData
})
.then(response => {
    if (response.status !== 200) {
        return throwError("Failed to add icon file", response);
    } else {
        return response.json();
    }
});

export const renameIcon = (oldName: string, newName: string) => fetch(getEndpointUrl(`/icon/${oldName}`), {
    method: "PATCH",
    headers: {
        "Content-Type": "application/json; charset=utf-8"
    },
    credentials: "include",
    body: JSON.stringify({name: newName})
})
.then(response => {
    if (response.status !== 204) {
        return throwError("Failed to rename icon", response);
    }
});

export const deleteIcon = (iconName: string) => fetch(getEndpointUrl(`/icon/${iconName}`), {
    method: "DELETE",
    credentials: "include"
})
.then(response => {
    if (response.status !== 204) {
        return throwError("Failed to delete icon", response);
    }
});

export const addIconfile = (iconName: string, format: string, size: string, formData: FormData) => fetch(
    getEndpointUrl(`/icon/${iconName}/format/${format}/size/${size}`),
    {
        method: "POST",
        credentials: "include",
        body: formData
    }
)
.then(response => {
    if (response.status !== 201) {
        return throwError("Failed to add icon-file", response);
    }
});

export const deleteIconfile = (iconfilePath: string) => fetch(getEndpointUrl(iconfilePath), {
    method: "DELETE",
    credentials: "include"
})
.then(response => {
    if (response.status !== 204) {
        return throwError("Failed to delete icon file", response);
    }
});

export const getIconfileType = (iconfile: IconfileDescriptor) => `${iconfile.format}@${iconfile.size}`;

const ip2ipwu = (iconPath: IconPath) => ({
    format: iconPath.format,
    size: iconPath.size,
    path: iconPath.path,
    url: getEndpointUrl(iconPath.path)
});

export const createIconfileList: (iconPaths: Set<IconPath>) => List<IconPathWithUrl>
= iconPaths => iconPaths.map(iconPath => ip2ipwu(iconPath)).toList();

export const preferredIconfileType: (icon: IconDescriptor) => IconPath
= icon => {
    // If the icon has SVG format, prefer that
    const svgFiles = icon.paths.filter(iconfile => iconfile.format === "svg");
    return svgFiles.size > 0
        ? svgFiles.first()
        : icon.paths.first();
};

export const urlOfIconfile = (icon: IconDescriptor, iconfileType: IconPath) => {
    const sameIconfileTypeFilter: (iconPath: IconPath) => boolean
        = iconPath => iconPath.format === iconfileType.format && iconPath.size === iconfileType.size;
    const icnPath: IconPath = icon.paths.filter(sameIconfileTypeFilter).first();
    if (!icnPath) {
        throw new Error(`${iconfileType} not found in icon ${icon.name}`);
    }
    return getEndpointUrl(icnPath.path);
};

export const preferredIconfileUrl = (icon: IconDescriptor) => urlOfIconfile(icon, preferredIconfileType(icon));

export const indexInIconfileListOfType = (iconfileList: List<IconfileDescriptor>, iconfileType: IconPath) =>
    iconfileList.findIndex(iconfile => iconfile.format === iconfileType.format && iconfile.size === iconfileType.size);

export const getTags: () => Promise<List<string>> = () => fetch(getEndpointUrl("/tag"), {
    method: "GET",
    credentials: "include"
})
.then(response => {
    if (response.status !== 200) {
        return throwError("Failed to retrieve tags", response);
    } else {
        return response.json();
    }
})
.then(
    (json: string[]) => List(json)
);

export const addTag: (iconName: string, tagText: string) => Promise<void>
= (iconName, tagText) => fetch(getEndpointUrl(`/icon/${iconName}/tag`), {
    method: "POST",
    headers: {
        "Content-Type": "application/json; charset=utf-8"
    },
    credentials: "include",
    body: JSON.stringify({tag: tagText})
})
.then(response => {
    if (response.status !== 200) {
        return throwError("Failed to add tag", response);
    }
});

export const removeTag: (iconName: string, tag: string) => Promise<void>
= (iconName, tag) => fetch(getEndpointUrl(`/icon/${iconName}/tag/${tag}`), {
    method: "DELETE",
    headers: {
        "Content-Type": "application/json; charset=utf-8"
    },
    credentials: "include"
})
.then(response => {
    if (response.status !== 200) {
        return throwError("Failed to remove tag", response);
    }
});
