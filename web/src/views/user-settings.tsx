import * as React from "react";

import { logout } from "../services/user";
import AccountCircleIcon from "@mui/icons-material/AccountCircle";
import { IconButton, Menu, MenuItem } from "@mui/material";
import PopupState, { bindMenu, bindTrigger } from "material-ui-popup-state";

interface UserSettingsProps {
    username: string;
		idPlogoutUrl: string;
}

export class UserSettings extends React.Component<UserSettingsProps, never> {
    constructor(props: UserSettingsProps) {
        super(props);
    }

    public render() {
			return <div className="user-area">
				<PopupState variant="popover" popupId="demo-popup-menu">
					{(popupState) => (
						<React.Fragment>
							<IconButton color="secondary" {...bindTrigger(popupState)}>
								<AccountCircleIcon/>
							</IconButton>
							<Menu {...bindMenu(popupState)}>
								<MenuItem onClick={() => {
									popupState.close();
									logout(this.props.idPlogoutUrl);
								}}>Logout</MenuItem>
							</Menu>
						</React.Fragment>
					)}
				</PopupState>
		</div>;
	}
}
