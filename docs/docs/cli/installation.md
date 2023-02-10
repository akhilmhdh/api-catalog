# Installation

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
<TabItem value="mac" label="MacOS" default>
Use <a>brew</a> package manaager
<br/>
<br/>

```bash
brew install one-platform/apic
```

</TabItem>
<TabItem value="alpine" label="Alpine" >
We use cloudsmith to distribute packages
<br/>
<br/>

```bash
curl -1sLf \
  'https://dl.cloudsmith.io/public/OWNER/REPOSITORY/cfg/setup/bash.deb.sh' \
  | sudo bash
```

</TabItem>
<TabItem value="deb" label="Debian/Ubuntu">
We use cloudsmith to distribute packages
<br/>
<br/>

```bash
curl -1sLf \
  'https://dl.cloudsmith.io/public/OWNER/REPOSITORY/cfg/setup/bash.deb.sh' \
  | sudo bash
```

</TabItem>
</Tabs>

To make sure it's installed correctly you can run

```bash
apic help
```

### Run against an OpenAPI

Let's run against [petstore api](https://google.com)

```bash
apic run --schema https://petstore3.swagger.io/api/v3/openapi.json
```
