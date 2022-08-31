import React, { useEffect } from "react";
import { useState } from "react";
import { IconCell } from "./icon-cell";
import { describeAllIcons, IconDescriptor, deleteIcon } from "../../services/icon";
import { AppSettgins } from "../app-settings";
import { UserSettings } from "../user-settings";
import { hasAddIconPrivilege, hasUpdateIconPrivilege } from "../../services/user";
import AddIcon from "@mui/icons-material/Add";
import SearchIcon from "@mui/icons-material/Search";

import "./icon-list.styl";
import { useSelector } from "react-redux";
import { IconRepoState } from "../../state/reducers/root-reducer";
import { IconButton } from "@mui/material";
import { IconDetailsDialog } from "./icon-details-dialog";
import { useReporter } from "../../services/app-messages";

export const IconList = (): JSX.Element => {

	const settings = useSelector((state: IconRepoState) => state.app);

	const { reportError, reportInfo } = useReporter();

	const detailsDialogForCreate = false;

	const [icons, setIcons] = useState<IconDescriptor[]>([]);
	const [searchQuery, setSearchQuery] = useState("");
	const [selectedIcon, setSelectedIcon] = useState(null);
	const [iconDetailDialogVisible, setIconDetailDialogVisible] = useState(false);

	const getIcons = () => {
		return describeAllIcons()
		.then(
			icons => setIcons(icons),
			error => { throw error; }
		);
	};

	
	useEffect(() => {
		getIcons();
	}, []);

	if (!settings?.appInfo || !settings?.userInfo) {
		return null;
	}

	const filteredIcons = () => {
		if (searchQuery === "") {
			return icons;
		} else {
			return icons.filter(icon => {
				return icon.name.toLowerCase().indexOf(searchQuery.toLowerCase()) !== -1;
			});
		}
	};
		
	const handleIconUpdate = async (iconName: string) => {
		await getIcons();
		if (iconName) {
			setSelectedIcon(icons.find(icon => icon.name === iconName));
		} else {
			setSelectedIcon(icons?.[0]);
		}
	};
	
	const handleIconDelete = (iconName: string) =>
		deleteIcon(iconName)
		.then(
			() => {
				reportInfo(`Icon ${iconName} removed`);
				setIcons(icons.filter(i => i.name !== selectedIcon.name));
				setSelectedIcon(null);
				setIconDetailDialogVisible(false);
			},
			err => reportError(err.message)
		);

	return <div>
		<header className="top-header">
		<div className="inner-wrapper">
			<div className="branding">
					<AppSettgins versionInfo = {settings.appInfo.versionInfo} />
					<div className="app-description">
						<span>{settings.appInfo.appDescription}</span>
					</div>
			</div>
			<div className="right-control-group">
				<div className="search">
					<div className="search-input-wrapper">
						<IconButton className="search-button"><SearchIcon/></IconButton>
						<input type="text" className="search-input"
							value={searchQuery}
							onChange={
								event => {
									const newValue = event.target.value;
									setSearchQuery(newValue);
								}
							}
						/>
					</div>
				</div>
				<UserSettings username={settings.userInfo.username}/>
			</div>
		</div>
		</header>

		<div className="action-bar">
		{
			hasAddIconPrivilege(settings.userInfo)
			? <div className="add-icon">
					<IconButton onClick={() => {
						setSelectedIcon(undefined);
						setIconDetailDialogVisible(true);
					}}><AddIcon/></IconButton>
				</div>
			: null
		}
		</div>

		{
			iconDetailDialogVisible
			? 
			<IconDetailsDialog
				username={settings.userInfo.username}
				isOpen={iconDetailDialogVisible}
				iconDescriptor={selectedIcon}
				handleIconUpdate={iconName => handleIconUpdate(iconName)}
				handleIconDelete={iconName => handleIconDelete(iconName)}
				requestClose={() => setIconDetailDialogVisible(false)}
				editable={hasUpdateIconPrivilege(settings.userInfo)}
				startInEdit={detailsDialogForCreate}
			/>
			: null
		}

		<section className="inner-wrapper icon-grid">
			{filteredIcons().map((icon, key) =>
				<div key = {key} className="grid-cell">
					<IconCell icon={icon} reqestDetails = {
						() => {
							setSelectedIcon(icon);
							setIconDetailDialogVisible(true);
						}
					}/>
				</div>
			)}
		</section>

	</div>;
};
