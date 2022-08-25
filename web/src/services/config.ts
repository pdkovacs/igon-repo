import { getData } from "./fetch-utils";

export const iconfileTypes: { [key: string]: string[] } = Object.freeze({
	svg: [
		"18px", "24px", "48px"
	],
	png: [
		"18dp", "24dp", "36dp", "48dp", "144dp"
	]
});

export interface VersionInfo {
	version: string;
	commit: string;
}

export interface AppInfo {
	versionInfo: VersionInfo;
	appDescription: string;
}

const fetchAppInfo: () => Promise<AppInfo> = () => getData("/app-info", 200);

export const fetchConfig = () => fetchAppInfo();

export const defaultTypeForFile = (fileName: string) => {
	const formats = Object.keys(iconfileTypes);
	const filenameExtension = fileName.split(".").pop();
	let format = null;
	let size = null;
	if (formats.includes(filenameExtension)) {
		format = filenameExtension;
		size = iconfileTypes[filenameExtension][0];
	}
	return {
		format,
		size
	};
};
