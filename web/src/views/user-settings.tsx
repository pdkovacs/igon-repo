import * as React from "react";
import { Icon, MenuItem } from "@blueprintjs/core";
import { Select, ItemRenderer } from "@blueprintjs/select";

import "./user-settings.scss";
import { logout } from "../services/user";

interface UserSettingsProps {
    username: string;
}

const ActionSelect = Select.ofType<string>();
const actionItemRenderer: ItemRenderer<string>
    = (item, {handleClick, index}) => <MenuItem key={index} text={item} onClick={handleClick}/>;

export class UserSettings extends React.Component<UserSettingsProps, {}> {
    constructor(props: UserSettingsProps) {
        super(props);
    }

    public render() {
        return <div className="user-area">
            <div className="account-settings-head">
                <div className="account-button">
                    <Icon className="account-button-element" icon="user"/>
                    <ActionSelect className="account-button-element"
                            filterable={false}
                            items={["Logout"]}
                            itemRenderer={actionItemRenderer}
                            onItemSelect={logout}>
                        {this.props.username}
                    </ActionSelect>
                </div>
            </div>
        </div>;
    }
}
