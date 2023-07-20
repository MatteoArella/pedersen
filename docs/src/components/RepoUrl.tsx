import React, {PropsWithChildren} from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

export interface RepoUrlProps {
  readonly name: string;
}

export default function RepoUrl(props: RepoUrlProps): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  const url = siteConfig.customFields!.repoUrl as string || ""
  return (
    <a href={url} target="_blank">{props.name}</a>
  );
}
