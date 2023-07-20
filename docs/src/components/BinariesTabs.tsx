import React, {PropsWithChildren} from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';
import CodeBlockLine from '@theme/CodeBlock/Line';

export interface VariantProps {
  readonly value: string;
  readonly label: string;
  readonly binary: string;
  readonly powershell: boolean;
  readonly default?: boolean;
}

export function Variant(props: VariantProps) {}

function RenderVariant(props: VariantProps): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  if (props.powershell) {
    return (
      <TabItem value={props.value} label={props.label} default={props.default}>
    <CodeBlock className="shell" language="powershell" title="PowerShell">
    {`$VERSION = "${siteConfig.customFields.version}"; \`
$BINARY = "${props.binary}"; \`
New-Item -ItemType Directory -Force -Path $env:ProgramFiles\\Pedersen\\bin; \`
Invoke-WebRequest -OutFile $env:ProgramFiles\\Pedersen\\bin\\pedersen https://github.com/matteoarella/pedersen/releases/download/$VERSION/$BINARY`}
    </CodeBlock>
    <br/>
    Then add <CodeBlock className="shell" language="powershell">$env:ProgramFiles\Pedersen\bin</CodeBlock> to your 
    <CodeBlock className="shell" language="powershell">$env:Path</CodeBlock> system variable.
    </TabItem>
    );
  }
  return (
    <TabItem value={props.value} label={props.label} default={props.default}>
  <CodeBlock className="shell" language="bash" title="sh / bash / zsh">
  {`VERSION=${siteConfig.customFields.version} \\
BINARY=${props.binary} \\
curl -o /bin/pedersen https://github.com/matteoarella/pedersen/releases/download/$VERSION/$BINARY && \\
chmod +x /bin/pedersen`}
  </CodeBlock>
  </TabItem>
  );
}

export interface PlatformProps {
  readonly value: string;
  readonly label: string;
  readonly default?: boolean;
}

export function Platform(props: PropsWithChildren<PlatformProps>) {}

function RenderPlatform(platform: PropsWithChildren<PlatformProps>): JSX.Element {
  return (
    <TabItem 
      default={platform.default}
      value={platform.value} label={platform.label}>
      <Tabs>{
        React.Children.map(platform.children, (variant, _) => {
          const item = variant as React.ReactElement<VariantProps>;
          const props: VariantProps = {
            powershell: platform.value === 'windows',
            ...item.props,
          }
          return RenderVariant(props)
        })
      }</Tabs>
    </TabItem>
  )
}

export function BinariesTabs({children}: {
  children: React.ReactElement<PropsWithChildren<PlatformProps>>
}): JSX.Element {
  return (<Tabs>{
    React.Children.map(children, (platform, _) => {
      return RenderPlatform(platform.props)
    })
  }</Tabs>)
};
