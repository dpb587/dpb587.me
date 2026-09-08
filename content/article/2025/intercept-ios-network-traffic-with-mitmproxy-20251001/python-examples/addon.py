import mitmproxy

class Addon:
  def response(self, flow: mitmproxy.http.HTTPFlow):
    print(f"{flow.request.method} {flow.request.url}")
    print(f"--> {flow.response.http_version} {flow.response.status_code} {flow.response.reason}")

addons = [Addon()]
