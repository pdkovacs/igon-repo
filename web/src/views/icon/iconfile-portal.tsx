import * as React from "react";
import { SelectFileToUpload } from "./select-file-to-upload";
import { ingestIconfile, createIcon, IngestedIconfileDTO } from "../../services/icon";
import { showErrorMessage } from "../../services/toasters";

import "./iconfile-portal.scss";

interface IconfilePortalProps {
    iconName: string;
    imageUrl: string;
    handleFileUpload: (uploadedFile: IngestedIconfileDTO) => void;
}

const uploadIconfile = (file: File, props: IconfilePortalProps) => {
    const fileName = file.name;
    const formData = new FormData();
    formData.append("file", file, fileName);

    let request: Promise<IngestedIconfileDTO>;
    if (props.iconName) {
        request = ingestIconfile(props.iconName, formData);
    } else {
        formData.append("name", fileName.replace(/(.*)\.[^.]*$/, "$1"));
        request = createIcon(formData);
    }

    request.then(
        iconfileMetadata => {
            props.handleFileUpload(iconfileMetadata);
        },
        error => showErrorMessage(error)
    )
    .catch(error => showErrorMessage(error));
};

const hasImageUrl = (props: IconfilePortalProps) => typeof props.imageUrl !== "undefined";
const portContent = (props: IconfilePortalProps) =>
    hasImageUrl(props)
        ? <img src={props.imageUrl}/>
        : <SelectFileToUpload handleSelectedFile={file => uploadIconfile(file, props)}/>;

export const IconfilePortal = (props: IconfilePortalProps) =>
<div className="iconfile-portal">
    {portContent(props)}
</div>;
