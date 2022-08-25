import { Autocomplete, Chip, TextField } from "@mui/material";
import React from "react";

import "./tag-collection.styl";

export interface TagForSelection {
	readonly text: string;
	readonly toBeCreated: boolean;
}

export interface TagCollectionProps {
	readonly selectedTags: string[];
	readonly selectionChangeRequest?: (tagText: string) => void;
	readonly allTags?: string[];
	readonly tagAdditionRequest?: (tagText: string) => void;
	readonly tagRemovalRequest?: (tagText: string) => void;
}

export const TagCollection = (props: TagCollectionProps) => {

	const handleOnClick = (tagText: string) => {
		if (props.selectionChangeRequest) {
			props.selectionChangeRequest(tagText);
		}
	};

	const createOnRemoveHandler = (tagText: string) => {
		return props.tagRemovalRequest
			? () => props.tagRemovalRequest(tagText)
			: undefined;
	};

	const createTag = (tagText: string) => {
		return <Chip className="tag-collection-item"
			label={tagText}
			key={tagText}
			onClick={() => handleOnClick(tagText)}
			onDelete={createOnRemoveHandler(tagText)}
		/>;
	};

	return <div className="tag-collection">
		{
			props.tagAdditionRequest
				? <Autocomplete
						multiple
						id="tags-filled"
						options={props.allTags.filter(tag => !props.selectedTags.includes(tag))}
						value={props.selectedTags}
						freeSolo
						renderTags={(value: readonly string[]) => value.map((option: string) => createTag(option))}
						renderInput={(params) => (
							<TextField
								{...params}
								variant="standard"
							/>
						)}
						onChange={(event: React.SyntheticEvent, value: string[]) => {
							props.tagAdditionRequest(value.filter(tag => !props.selectedTags.includes(tag))?.[0]);
						}}
					/>
				: props.selectedTags.map(t => createTag(t))
		}
	</div>;
};

