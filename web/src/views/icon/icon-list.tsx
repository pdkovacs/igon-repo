import * as React from "react";
import { IconCell } from "./icon-cell";
import { describeAllIcons, IconDescriptor, getTags } from "../../services/icon";
import { Set } from "immutable";
import { fetchConfig, AppInfo } from "../../services/config";
import { AppSettgins } from "../app-settings";
import { UserSettings } from "../user-settings";
import { UserInfo, fetchUserInfo, hasAddIconPrivilege, hasUpdateIconPrivilege } from "../../services/user";

import "./icon-list.scss";
import { Icon, Button, Intent } from "@blueprintjs/core";
import { IconDetailsDialog } from "./icon-details-dialog";

interface Settings {
    readonly appInfo: AppInfo;
    readonly userInfo: UserInfo;
}

const initialSettings: Settings = {
    appInfo: {
        versionInfo: {
            version: "No data",
            commit: "No data"
        },
        appDescription: "No data"
     },
     userInfo: {
         privileges: Set([]),
         username: "John Doe",
         authenticated: false
     }
};

interface IconListState {
    readonly settings: Settings;
    readonly icons: Set<IconDescriptor>;
    readonly searchQuery: string;
    readonly selectedIcon: IconDescriptor;
    readonly iconDetailDialogVisible: boolean;
    readonly allTags: Set<string>;
}

export class IconList extends React.Component<{}, IconListState> {

    private detailsDialogForCreate: boolean = false;

    constructor(props: {}) {
        super(props);
        this.state = {
            settings: { ...initialSettings },
            icons: Set([]),
            searchQuery: "",
            selectedIcon: null,
            iconDetailDialogVisible: false,
            allTags: Set([])
        };
    }

    public componentDidMount() {
        fetchConfig()
        .then(
            appInfo => this.setState(prevState => ({ ...prevState, settings: { ...prevState.settings, appInfo } }))
        );
        fetchUserInfo()
        .then(
            userInfo => this.setState(prevState => ({ ...prevState, settings: { ...prevState.settings, userInfo } } ))
        );
        describeAllIcons()
        .then(
            icons => this.setState(prevState => ({ ...prevState, icons })),
            error => { throw error; }
        );
        getTags()
        .then(
            tags => this.setState({allTags: Set(tags)})
        );
    }

    public render() {
        return <div>
            <header className="top-header">
            <div className="inner-wrapper">
                <div className="branding">
                    <AppSettgins versionInfo = {this.state.settings.appInfo.versionInfo} />
                    <div className="app-description">
                        <span>{this.state.settings.appInfo.appDescription}</span>
                    </div>
                </div>
                <div className="right-control-group">
                    <div className="search">
                        <div className="search-input-wrapper">
                            <Icon className="search-icon" icon="search" iconSize={24}/>
                            <input type="text" className="search-input"
                                value={this.state.searchQuery}
                                onChange={
                                    event => {
                                        const newValue = event.target.value;
                                        this.setState(
                                            prevState => ({...prevState, searchQuery: newValue})
                                        );
                                    }
                                }
                            />
                        </div>
                    </div>
                    <UserSettings username={this.state.settings.userInfo.username}/>
                </div>
            </div>
            </header>

            <div className="action-bar">
                {this.actionBarContent()}
            </div>

            {this.iconDetailsDialog()}

            <section className="inner-wrapper icon-grid">
                {this.filteredIcons().toList().map((icon, key) =>
                    <div key = {key} className="grid-cell">
                        <IconCell icon = {icon} reqestDetails = {
                            () => this.setState({selectedIcon: icon, iconDetailDialogVisible: true})
                        }/>
                    </div>
                )}
            </section>

        </div>;
    }

    private filteredIcons() {
        if (this.state.searchQuery === "") {
            return this.state.icons;
        } else {
            return this.state.icons.filter(icon => {
                return icon.name.toLowerCase().indexOf(this.state.searchQuery.toLowerCase()) !== -1;
            });
        }
    }

    private actionBarContent() {
        if (hasAddIconPrivilege(this.state.settings.userInfo)) {
            return <div className="add-icon">
                <Button intent={Intent.PRIMARY} icon="plus" onClick={() => this.openCreateIconDialog()}>ADD NEW</Button>
            </div>;
        } else {
            return null;
        }
    }

    private openCreateIconDialog() {
        this.setState({selectedIcon: undefined, iconDetailDialogVisible: true});
    }

    private iconDetailsDialog() {
        if (this.state.iconDetailDialogVisible) {
            return <IconDetailsDialog
                username={this.state.settings.userInfo.username}
                isOpen={this.state.iconDetailDialogVisible}
                iconDescriptor={this.state.selectedIcon}
                handleIconUpdate={icon => this.handleIconUpdate(icon)}
                handleIconDelete={() => this.handleIconDelete()}
                requestClose={() => this.setState({iconDetailDialogVisible: false})}
                editable={hasUpdateIconPrivilege(this.state.settings.userInfo)}
                startInEdit={this.detailsDialogForCreate}
                tags={this.state.allTags}/>;
        } else {
            return null;
        }
    }

    private handleIconUpdate(icon: IconDescriptor) {
        const tmp: Set<IconDescriptor> = this.state.selectedIcon // update or creation
            ? this.state.icons.filter(i => i.name !== this.state.selectedIcon.name)
            : this.state.icons;
        this.setState({
            icons: tmp.add(icon),
            selectedIcon: icon,
            allTags: this.state.allTags.union(icon.tags)
        });
    }

    private handleIconDelete() {
        this.setState({
            icons: this.state.icons.filter(i => i.name !== this.state.selectedIcon.name),
            selectedIcon: null
        });
    }
}
