# Supported and Planned Features

The table below contains the current state of the Kusk Gateway features.

Features marked with the ✅ are implemented, not marked are planned to be implemented.

For the features currently on the roadmap please see [Kusk-gateway milestones](https://github.com/kubeshop/kusk-gateway/milestones).

<style type="text/css">
    .tg {
        border-collapse: collapse;
        border-spacing: 0;
    }

    .tg td {
        border-color: black;
        border-style: solid;
        border-width: 1px;
        font-family: Arial, sans-serif;
        font-size: 14px;
        overflow: hidden;
        padding: 10px 5px;
        word-break: normal;
    }

    .tg th {
        border-color: black;
        border-style: solid;
        border-width: 1px;
        font-family: Arial, sans-serif;
        font-size: 14px;
        font-weight: normal;
        text-align: center;
        overflow: hidden;
        padding: 10px 5px;
        word-break: normal;
    }

    .tg .tg-0pky {
        border-color: inherit;
        text-align: left;
        vertical-align: top
    }
</style>
<table class="tg">
    <tbody>
        <tr>
            <th class="tg-0pky">
                <p><strong>Categories</strong></p>
            </th>
            <th class="tg-0pky">
                <p><strong>Feature/Description with the implementation status</strong></p>
            </th>
            <th class="tg-0pky">
                <p><strong>Comments</strong></p>
            </th>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Basic routing</p>
            </td>
            <td class="tg-0pky">
                <p></p>Configure the routing by: </p>
                <ul>
                    <li>
                        <p>host ✅</p>
                    </li>
                    <li>
                        <p>exact path ✅</p>
                    </li>
                    <li>
                        <p>path with a regexp ✅</p>
                    </li>
                    <li>
                        <p>path with a prefix ✅</p>
                    </li>
                </ul>
            </td>
            <td class="tg-0pky">
                <p> </p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Basic routing with OpenAPI</p>
            </td>
            <td class="tg-0pky">
                <p>Configure the basic routing with OpenAPI + x-kusk extension ✅</p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>HTTP redirect</p>
            </td>
            <td class="tg-0pky">
                <p>HTTP redirect support, with dynamic path rewrites ✅</p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Direct HTTP response</p>
            </td>
            <td class="tg-0pky">
                <p>Respond with HTTP code without sending to the upstream </p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>HTTP path manipulation</p>
            </td>
            <td class="tg-0pky">
                <ul>
                    <li>
                        Prepend and strip HTTP path prefix for the OpenAPI paths ✅
                    </li>
                    <li>
                        Rewrite paths when sending to the upstream (backend) ✅
                    </li>
                </ul>
            </td>
            <td class="tg-0pky">
                <p> </p>
                <p></p>
                <p>
                </p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>HTTP sticky sessions</p>
            </td>
            <td class="tg-0pky">
                <p>Binding the client session to the same upstream host by IP address </p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>HTTP headers manipulation</p>
            </td>
            <td class="tg-0pky">
                <p>Inserting/removing headers when communicating with the upstream </p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>HTTP Compression</p>
            </td>
            <td class="tg-0pky">
                <p>Gzip/Bzip/Brotli headers/body compression </p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>TLS</p>
            </td>
            <td class="tg-0pky">
                <ul>
                    <li>Static (externally deployed) certificates ✅ </li>
                    <li>LetsEncrypt (ACME) dynamically configured certificates </li>
                </ul>
            </td>
            <td class="tg-0pky"> <p></p> </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>CORS</p>
            </td>
            <td class="tg-0pky">
                <p>CORS support ✅</p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Websockets</p>
            </td>
            <td class="tg-0pky">
                <p>Websockets support ✅</p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Quality of the service<br></p>
                <p></p>
            </td>
            <td class="tg-0pky">
                <ul>
                    <li>
                        <p>Retries on 50x code ✅</p>
                    </li>
                    <li>
                        <p>Request timeouts, idle timeouts ✅
                        </p>
                    </li>
                    <li>
                        <p>Rate limiting </p>
                    </li>
                    <li>
                        <p>Cirquit breaker </p>
                    </li>
                </ul>
            </td>
            <td class="tg-0pky">
                <p>​</p>
                <p></p>
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Advanced routing</p>
            </td>
            <td class="tg-0pky">
                <ul>
                    <li>
                        <p>Traffic mirroring (sending the requests both to the actual upstream and mirror to some other sink, e.g. staging) </p>
                    </li>
                    <li>
                        <p>Traffic splitting (sending to 2 or more services at the same time) </p>
                    </li>
                </ul>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Validation</p>
            </td>
            <td class="tg-0pky">
                <p>Validation of requests and responses using OpenAPI definition </p>
            </td>
            <td class="tg-0pky">
                <p>Partially: the validation of requests are in alpha mode</p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Mocking</p>
            </td>
            <td class="tg-0pky">
                <p>Mocking endpoints using OpenAPI definition </p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Visibility: Dashboard</p>
            </td>
            <td class="tg-0pky">
                <p>User friendly management portal </p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Visibility: Logging</p>
            </td>
            <td class="tg-0pky">
                <p>Ability to collect, aggregate and analyze access and gateway logs </p>
            </td>
            <td class="tg-0pky">
                <p>Partially: enabled stdout access logging that can be used by third party tools like Fluentd.</p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Visibility: Tracing</p>
            </td>
            <td class="tg-0pky">
                <p>Ability to trigger requests tracing </p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Visibility: Metrics</p>
            </td>
            <td class="tg-0pky">
                <p>Ability to collect and analyze traffic and gateway metrics </p>
            </td>
            <td class="tg-0pky">
                <p>Partially: basic HTTP requests Prometheus metrics could be collected from Envoy</p>
            </td>
        </tr>
        <tr>
            <td class="tg-0pky">
                <p>Authentication</p>
            </td>
            <td class="tg-0pky">
                <p>Authentication schemes support with <a href="https://swagger.io/docs/specification/authentication/">OpenAPI security</a> and separately in StaticRoute</p>
            </td>
            <td class="tg-0pky">
                <p></p>
            </td>
        </tr>
    </tbody>
</table>
