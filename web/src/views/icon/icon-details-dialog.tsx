import React, { useEffect, useState } from "react";

import {
	IconDescriptor,
	createIconfileList,
	preferredIconfileType,
	IconPathWithUrl,
	IngestedIconfileDTO,
	deleteIconfile,
	IconfileDescriptor,
	getIconfileType,
	addTag,
	removeTag,
	getTags
} from "../../services/icon";
import { TagCollection } from "../tag-collection";
import { renderMapAsTable } from "../layout-util";
import getUrl from "../../services/url";
import { IconfilePortal } from "./iconfile-portal";

import "./icon-details-dialog.styl";
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton } from "@mui/material";
import VisibilityIcon from "@mui/icons-material/Visibility";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import { isNil } from "lodash";
import { useReporters } from "../../utils/use-reporters";

interface IconDetailsDialogProps {
	readonly username: string;
	readonly isOpen: boolean;
	readonly iconDescriptor: IconDescriptor;
	readonly handleIconUpdate: (iconName: string) => void;
	readonly handleIconDelete: (iconName: string) => void;
	readonly requestClose: () => void;
	readonly editable: boolean;
	readonly startInEdit: boolean;
}

interface IconDetailsTitleBarControlProps {
	readonly className: string;
	readonly children: JSX.Element;
	readonly toShow: boolean;
	readonly action: () => void;
}
const IconDetailsTitleBarControl = (props: IconDetailsTitleBarControlProps) => {
	return props.toShow
		? <IconButton className={`title-control ${props.className}`} onClick={props.action}>
			{props.children}
		</IconButton>
		: null;
};

interface Action {
	readonly view: () => void;
	readonly edit: () => void;
	readonly delete: () => void;
}
interface IconDetailsTitleBarProps {
	readonly iconName: string;
	readonly inEdit: boolean;
	readonly editable: boolean;
	readonly action: Action;
}

const IconDetailsTitleBar = (props: IconDetailsTitleBarProps) => {
	return <div className="title-bar">
		<span>{props.iconName}</span>
		<IconDetailsTitleBarControl className="view-icon-button" toShow={props.editable && props.inEdit} action={props.action.view}>
			<VisibilityIcon />
		</IconDetailsTitleBarControl>
		<IconDetailsTitleBarControl className="edit-icon-button" toShow={props.editable && !props.inEdit} action={props.action.edit}>
			<EditIcon />
		</IconDetailsTitleBarControl>
		<IconDetailsTitleBarControl className="delete-icon-button" toShow={props.editable && true} action={props.action.delete}>
			<DeleteIcon />
		</IconDetailsTitleBarControl>
	</div>;
};

const iconfileType = (iconfile: IconfileDescriptor) => {
	return `${iconfile.format}@${iconfile.size}`;
};

export const IconDetailsDialog = (props: IconDetailsDialogProps) => {

	const initialIconfileSelection = (): IconfileDescriptor => {
		if (!props.iconDescriptor) {
			return null;
		}
		const icon = props.iconDescriptor;
		return preferredIconfileType(icon);
	};

	const [allTags, setAllTags] = useState<string[]>(null);

	const [iconName, setIconName] = useState<string>(props.iconDescriptor ? props.iconDescriptor.name : null);
	const [iconTags, setIconTags] = useState<string[]>(props.iconDescriptor ? props.iconDescriptor.tags : []);
	const [modifiedBy, setModifiedBy] = useState<string>(props.iconDescriptor ? props.iconDescriptor.modifiedBy : "<none>");
	const [inEdit, setInEdit] = useState(props.startInEdit);
	const [selectedIconfile, setSelectedIconfile] = useState<IconfileDescriptor>(initialIconfileSelection());
	const [previouslySelectedIconFile, setPreviouslySelectedIconFile] = useState<IconfileDescriptor>(null);

	const [iconfileList, setIconfileList] = useState<IconPathWithUrl[]>(props.iconDescriptor ? createIconfileList(props.iconDescriptor.paths) : []);

	const { reportError, reportInfo } = useReporters();

	useEffect(() => {
		getTags()
		.then(
			tags => setAllTags(tags)
		);
	}, []);

	const iconfileFormats = () => iconfileList.map(iconfile => iconfileType(iconfile));

	const findByDescriptorIn = (descriptor: IconfileDescriptor, fileList: IconPathWithUrl[]) =>
		fileList.find(iconfile => iconfile.format === descriptor.format && iconfile.size == descriptor.size);

	const pathOfSelectedIconfile = React.useMemo(() => {
		return isNil(selectedIconfile)
			? undefined
			: findByDescriptorIn(selectedIconfile, iconfileList)?.url;
	}, [iconfileList, selectedIconfile]);

	const addTagToIcon = (tagToAdd: string) => {
		addTag(iconName, tagToAdd)
			.then(
				() => {
					setAllTags(allTags.concat([tagToAdd]));
					setIconTags(iconTags.concat([tagToAdd]));
					props.handleIconUpdate(iconName);
				},
				error => reportError(error)
			);
	};

	const removeIconTag = (tag: string) => {
		removeTag(iconName, tag)
			.then(
				() => {
					setIconTags(iconTags.filter(iconTag => iconTag !== tag));
					props.handleIconUpdate(null);
				}
			);
	};

	const createTagList = () => {
		if (inEdit) {
			return <TagCollection
				selectedTags={iconTags}
				tagAdditionRequest={tagToAdd => addTagToIcon(tagToAdd)}
				tagRemovalRequest={index => removeIconTag(index)}
				allTags={allTags} />;
		} else {
			return <TagCollection
				selectedTags={iconTags}
				allTags={[]} />;
		}
	};

	const iconfileDeletionRequest = (fileType: string) => {
		const selectedFormat = iconfileList.find(iconfile => iconfileType(iconfile) === fileType);
		const newIconfileList = iconfileList.filter(iconfile => iconfileType(iconfile) !== fileType);
		deleteIconfile(selectedFormat.path)
			.then(
				() => {
					setIconfileList(newIconfileList);
					setSelectedIconfile(newIconfileList?.[0]);
					setModifiedBy(props.username);
					setInEdit(false);
					reportInfo(`Iconfile ${getIconfileType(selectedFormat)} removed`);
					props.handleIconUpdate(newIconfileList.length ? iconName : null);
					if (newIconfileList.length === 0) {
						props.requestClose();
					}
				},
				error => reportError(error.toString())
			);
	};

	const propertiesRow = () => {
		return <div className="properties-row">
			{
				renderMapAsTable({
					"Last modified by":
						<TagCollection
							key={1}
							selectedTags={[modifiedBy]}
							allTags={[]}
						/>,
					"Available formats":
						<TagCollection
							key={2}
							selectedTags={iconfileFormats()}
							selectionChangeRequest={type => setSelectedIconfile(iconfileList.find(iconfile => iconfileType(iconfile) === type))}
							tagRemovalRequest={
								inEdit
									? fileFormat => iconfileDeletionRequest(fileFormat)
									: undefined
							}
							allTags={[]}
						/>,
					"Tags":
						createTagList()
				})
			}
		</div>;
	};

	const viewIcon: () => void = () => {
		setSelectedIconfile(previouslySelectedIconFile);
		setInEdit(false);
	};

	const editIcon: () => void = () => {
		setInEdit(true);
		setPreviouslySelectedIconFile(selectedIconfile);
		setSelectedIconfile(null);
	};

	const deleteIcon: () => void = () => props.handleIconDelete(iconName);

	const handleIconfileUpload = (uploadedFile: IngestedIconfileDTO) => {
		const newIconfileList = iconfileList.concat([{ ...uploadedFile, url: getUrl(uploadedFile.path) }]);
		setIconName(uploadedFile.iconName);
		setIconfileList(newIconfileList);
		setSelectedIconfile(findByDescriptorIn(uploadedFile, newIconfileList));
		setModifiedBy(props.username);
		setInEdit(false);
		props.handleIconUpdate(uploadedFile.iconName);
		reportInfo(`Iconfile ${uploadedFile.iconName}@${uploadedFile.size}.${uploadedFile.format} added`);
		props.requestClose();
	};

	const downloadName = () => {
		return selectedIconfile
			? `${iconName}@${selectedIconfile.size}.${selectedIconfile.format}`
			: "";
	};

	return !isNil(allTags) && <Dialog className="icon-details-dialog" open={props.isOpen} >
		<DialogTitle>
			<IconDetailsTitleBar iconName={props.iconDescriptor?.name} editable={props.editable} inEdit={inEdit} action={{
				view: viewIcon,
				edit: editIcon,
				delete: deleteIcon
			}} />
		</DialogTitle>
		<DialogContent>
			<div className="icon-box-row">
				<IconfilePortal
					imageUrl={pathOfSelectedIconfile}
					iconName={iconName}
					handleFileUpload={uploadedFile => handleIconfileUpload(uploadedFile)}
					handleError={error => reportError(error.toString())}
				/>
				{propertiesRow()}
			</div>
		</DialogContent>
		<DialogActions>
			<Button
				href={pathOfSelectedIconfile}
				download={downloadName()}
				disabled={!pathOfSelectedIconfile}
			>
				Download
			</Button>
			<Button onClick={props.requestClose}>Close</Button>
		</DialogActions>
	</Dialog>;
};
