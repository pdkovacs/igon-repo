import getEndpointUrl from "./url";
import { deleteData, getData, patchData, postData } from "./fetch-utils";

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
    readonly paths: IconPath[];
    readonly tags: string[];
}

export const describeAllIcons: () => Promise<IconDescriptor[]> = () => getData("/icon", 200);
export const describeIcon = (iconName: string) => getData(`/icon/${iconName}`, 200);
export const createIcon: (formData: FormData) => Promise<IconDescriptor> = formData => postData("/icon", 201, null, formData, false);
export const renameIcon = (oldName: string, newName: string) => patchData(`/icon/${oldName}`, 204, null, {name: newName});
export const deleteIcon = (iconName: string) => deleteData(`/icon/${iconName}`, 204);

export const ingestIconfile: (iconName: string, formData: FormData) => Promise<IngestedIconfileDTO>
	= (iconName, formData) => postData(`/icon/${iconName}`, 200, null, formData, false);
export const deleteIconfile = (iconfilePath: string) => deleteData(iconfilePath, 204);

export const getIconfileType = (iconfile: IconfileDescriptor) => `${iconfile.format}@${iconfile.size}`;

const ip2ipwu = (iconPath: IconPath) => ({
    format: iconPath.format,
    size: iconPath.size,
    path: iconPath.path,
    url: getEndpointUrl(iconPath.path)
});

export const createIconfileList: (iconPaths: IconPath[]) => IconPathWithUrl[]
	= iconPaths => iconPaths.map(iconPath => ip2ipwu(iconPath));

export const preferredIconfileType: (icon: IconDescriptor) => IconPath
	= icon => {
			// If the icon has SVG format, prefer that
			const svgFiles = icon.paths.filter(iconfile => iconfile.format === "svg");
			return svgFiles.length > 0
					? svgFiles?.[0]
					: icon.paths?.[0];
	};

export const urlOfIconfile = (icon: IconDescriptor, iconfileType: IconPath) => {
    const sameIconfileTypeFilter: (iconPath: IconPath) => boolean
        = iconPath => iconPath.format === iconfileType.format && iconPath.size === iconfileType.size;
    const icnPath: IconPath = icon.paths.filter(sameIconfileTypeFilter)?.[0];
    if (!icnPath) {
        throw new Error(`${iconfileType} not found in icon ${icon.name}`);
    }
    return getEndpointUrl(icnPath.path);
};

export const preferredIconfileUrl = (icon: IconDescriptor) => urlOfIconfile(icon, preferredIconfileType(icon));

export const getTags: () => Promise<string[]> = () => getData("/tag", 200);
export const addTag: (iconName: string, tagText: string) => Promise<void>
	= (iconName, tagText) => postData(`/icon/${iconName}/tag`, 201, null, {tag: tagText});
export const removeTag: (iconName: string, tag: string) => Promise<void> = (iconName, tag) => deleteData<void, void>(`/icon/${iconName}/tag/${tag}`, 204);
