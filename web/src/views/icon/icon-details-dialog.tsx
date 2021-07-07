import { List, Map, Set } from "immutable";
import * as React from "react";
import { Dialog, Classes, Icon, IconName, AnchorButton } from "@blueprintjs/core";

import {
    IconDescriptor,
    createIconfileList,
    preferredIconfileType,
    indexInIconfileListOfType,
    IconPathWithUrl,
    IngestedIconfileDTO,
    deleteIconfile,
    deleteIcon,
    IconfileDescriptor,
    getIconfileType,
    addTag,
    removeTag} from "../../services/icon";
import { TagCollection } from "../tag-collection";
import { renderMapAsTable } from "../layout-util";
import getUrl from "../../services/url";
import { showSuccessMessage, showErrorMessage } from "../../services/toasters";
import { IconfilePortal } from "./iconfile-portal";

import "./icon-details-dialog.scss";

interface IconDetailsDialogProps {
    readonly username: string;
    readonly isOpen: boolean;
    readonly iconDescriptor: IconDescriptor;
    readonly handleIconUpdate: (icon: IconDescriptor) => void;
    readonly handleIconDelete: (iconName: string) => void;
    readonly requestClose: () => void;
    readonly editable: boolean;
    readonly startInEdit: boolean;
    readonly tags: Set<string>;
}

interface IconDetailsDialogState {
    readonly iconName: string;
    readonly modifiedBy: string;
    readonly iconfileList: List<IconPathWithUrl>;
    readonly iconTags: List<string>;
    readonly selectedIconfileIndex: number;
    readonly previouslySelectedIconFile: number;
    readonly inEdit: boolean;
    readonly allTags: Set<string>;
}

const staticDialogOptions = {
    autoFocus: true,
    canEscapeKeyClose: true,
    canOutsideClickClose: false,
    enforceFocus: true,
    hasBackdrop: false,
    usePortal: true,
    style: {
        width: "40%",
        backgroundColor: "white"
    }
};

export class IconDetailsDialog extends React.Component<IconDetailsDialogProps, IconDetailsDialogState> {
    constructor(props: IconDetailsDialogProps) {
        super(props);
        this.state = {
            iconName: props.iconDescriptor ? props.iconDescriptor.name : null,
            iconfileList: props.iconDescriptor ? createIconfileList(props.iconDescriptor.paths) : List.of(),
            iconTags: props.iconDescriptor ? props.iconDescriptor.tags.toList() : List.of(),
            selectedIconfileIndex: this.initialIconfileSelection(),
            modifiedBy: props.iconDescriptor ? props.iconDescriptor.modifiedBy : "<none>",
            previouslySelectedIconFile: -1,
            inEdit: props.startInEdit,
            allTags: props.tags
        };
    }

    public render() {
        return <Dialog
            {...staticDialogOptions}
            title={this.createTitleBar()}
            isOpen={this.props.isOpen}
            onClose={this.handleClose}>
            <div className={Classes.DIALOG_BODY}>
                <div className="icon-details-dialog">
                    <div className="icon-box-row">
                        <IconfilePortal
                            imageUrl={this.pathOfSelectedIconfile()}
                            iconName={this.state.iconName}
                            handleFileUpload={uploadedFile => this.handleIconfileUpload(uploadedFile)} />
                    </div>
                    <div className="properties-row">
                        {renderMapAsTable(Map.of(
                            "Last modified by",
                            <TagCollection
                                        tags={List.of(this.state.modifiedBy)}
                                        tagsAvailableForAddition={Set.of()}/>,
                            "Available formats",
                            <TagCollection
                                        tags={this.iconfileFormats()}
                                        selectedIndex={this.state.selectedIconfileIndex}
                                        selectionChangeRequest={i => this.setState({selectedIconfileIndex: i})}
                                        tagRemovalRequest={
                                            this.state.inEdit
                                                ? tagText => this.iconfileDeletionRequest(tagText)
                                                : undefined}
                                        tagsAvailableForAddition={Set.of()}/>,
                            "Tags",
                            this.createTagList()
                        ))}
                    </div>
                </div>
            </div>
            <div className={Classes.DIALOG_FOOTER}>
                <div className="dialog-footer-actions">
                    <AnchorButton
                            href={this.pathOfSelectedIconfile()}
                            download={this.downloadName()}
                            disabled={!this.pathOfSelectedIconfile()}>
                        Download
                    </AnchorButton>
                </div>
            </div>
        </Dialog>;
    }

    private pathOfSelectedIconfile() {
        return this.state.selectedIconfileIndex >= 0
        ? this.state.iconfileList.get(this.state.selectedIconfileIndex).url
        : undefined;
    }

    private downloadName() {
        const iconfile = this.state.iconfileList.get(this.state.selectedIconfileIndex);
        return iconfile
            ? `${this.state.iconName}@${iconfile.size}.${iconfile.format}`
            : "";
    }

    private initialIconfileSelection() {
        if (!this.props.iconDescriptor) {
            return -1;
        }
        const icon = this.props.iconDescriptor;
        const preferredType = preferredIconfileType(icon);
        return indexInIconfileListOfType(createIconfileList(icon.paths), preferredType);
    }

    private handleIconfileUpload(uploadedFile: IngestedIconfileDTO) {
        const newIconfileList = this.state.iconfileList.push({...uploadedFile, url: getUrl(uploadedFile.path)});
        const newState: IconDetailsDialogState = {
            iconName: uploadedFile.iconName,
            iconfileList: newIconfileList,
            iconTags: this.state.iconTags,
            selectedIconfileIndex: indexInIconfileListOfType(newIconfileList, uploadedFile),
            previouslySelectedIconFile: this.state.previouslySelectedIconFile,
            modifiedBy: this.props.username,
            inEdit: false,
            allTags: this.state.allTags
        };
        this.setState(newState);
        this.notifyOfUpdate();
        showSuccessMessage(`Iconfile ${this.getSelectedFormat()} added`);
    }

    private iconfileFormats() {
        return this.state.iconfileList.map(iconfile => this.iconfileType(iconfile));
    }

    private iconfileType(iconfile: IconfileDescriptor) {
        return `${iconfile.format}@${iconfile.size}`;
    }

    private getSelectedFormat() {
        return this.iconfileFormats().get(this.state.selectedIconfileIndex);
    }

    private viewIcon() {
        this.setState({
            selectedIconfileIndex: this.state.previouslySelectedIconFile,
            inEdit: false
        });
    }

    private editIcon() {
        this.setState({
            inEdit: true,
            previouslySelectedIconFile: this.state.selectedIconfileIndex,
            selectedIconfileIndex: -1
        });
    }

    private iconfileDeletionRequest(indexInFileList: number) {
        const selectedFormat = this.state.iconfileList.get(indexInFileList);
        deleteIconfile(selectedFormat.path)
        .then(
            () => {
                this.setState({
                    iconfileList: this.state.iconfileList.filter((file, i) => i !== indexInFileList),
                    selectedIconfileIndex: 0,
                    modifiedBy: this.props.username,
                    inEdit: false
                });
                showSuccessMessage(`Iconfile ${getIconfileType(selectedFormat)} removed`);
                this.notifyOfUpdate();
            },
            error => showErrorMessage(error)
        )
        .catch(error => showErrorMessage(error));
    }

    private deleteIcon() {
        deleteIcon(this.state.iconName)
        .then(
            () => {
                showSuccessMessage(`Icon ${this.state.iconName} removed`);
                this.props.handleIconDelete(this.state.iconName);
                this.props.requestClose();
            },
            error => showErrorMessage(error)
        )
        .catch(error => showErrorMessage(error));
    }

    private notifyOfUpdate() {
        this.props.handleIconUpdate({
            name: this.state.iconName,
            modifiedBy: this.state.modifiedBy,
            paths: this.state.iconfileList.toSet(),
            tags: this.state.iconTags.toSet()
        });
    }

    private createTitleBar() {
        return <div className="title-bar">
                    <span>{this.state.iconName}</span>
                    {this.createTitleBarControl("view-icon-button", "eye-open",
                                                this.state.inEdit, () => this.viewIcon())}
                    {this.createTitleBarControl("edit-icon-button", "edit",
                                                !this.state.inEdit, () => this.editIcon())}
                    {this.createTitleBarControl("delete-icon-button", "trash", true, () => this.deleteIcon())}
                </div>;
    }

    private createTitleBarControl(className: string, iconName: IconName, toShow: boolean, action: () => void) {
        return this.props.editable && toShow
            ? <Icon className={`title-control ${className}`} icon={iconName} onClick={action}/>
            : null;
    }

    private createTagList() {
        if (this.state.inEdit) {
            return <TagCollection
                    tags={this.state.iconTags}
                    selectedIndex={0}
                    tagAdditionRequest={tagToAdd => this.addTagToIcon(tagToAdd)}
                    tagRemovalRequest={index => this.removeTag(index)}
                    tagsAvailableForAddition={this.state.allTags}/>;
        } else {
            return <TagCollection
                tags={this.state.iconTags}
                selectedIndex={0}
                tagsAvailableForAddition={Set.of()}/>;
        }
    }

    private addTagToIcon(tagToAdd: string) {
        addTag(this.state.iconName, tagToAdd)
        .then(
            () => {
                this.setState({
                    allTags: this.state.allTags.add(tagToAdd),
                    iconTags: this.state.iconTags.push(tagToAdd)
                });
                this.notifyOfUpdate();
            },
            error => showErrorMessage(error)
        )
        .catch(error => showErrorMessage(error));
    }

    private removeTag(index: number) {
        const tagText: string = this.state.iconTags.get(index);
        removeTag(this.state.iconName, tagText)
        .then(
            () => {
                this.setState({iconTags: this.state.iconTags.remove(index)});
                this.notifyOfUpdate();
            }
        );
    }

    private handleClose = () => this.props.requestClose();

}
