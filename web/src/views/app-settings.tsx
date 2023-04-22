import { Button, Dialog, DialogActions, DialogContent } from "@mui/material";
import * as React from "react";
import { VersionInfo } from "../services/config";
import "./app-settings.styl";

const dialogOptions = {
    autoFocus: true,
    canEscapeKeyClose: true,
    canOutsideClickClose: true,
    enforceFocus: true,
    hasBackdrop: true,
    title: "About \"Icons\"",
    usePortal: true
};

export class AppSettgins extends React.Component<{versionInfo: VersionInfo}, {isOpen: boolean}> {

	constructor(props: {versionInfo: VersionInfo}) {
		super(props);
		this.state = { isOpen: false };
	}

	public render() {
		return <div>
			<h1 className="app-settings">
				<a onClick={this.handleOpen}>Icons</a>
			</h1>
			<Dialog
				{...dialogOptions}
				open={this.state.isOpen}
				onClose={this.handleClose}>
				<DialogContent className="about-dialog">
					<table>
						<tbody>
							<tr>
								<td>Version:</td><td>{this.props.versionInfo.version}</td>
							</tr>
							<tr>
								<td>Commit ID:</td><td>{this.props.versionInfo.commit}</td>
							</tr>
							<tr>
								<td>Build time:</td><td>{this.props.versionInfo.buildTime}</td>
							</tr>
					</tbody>
					</table>
				</DialogContent>
				<DialogActions>
					<Button onClick={this.handleClose}>Close</Button>
				</DialogActions>
			</Dialog>
		</div>;
	}

	private handleOpen = () => this.setState({ isOpen: true });
	private handleClose = () => this.setState({ isOpen: false });
}
