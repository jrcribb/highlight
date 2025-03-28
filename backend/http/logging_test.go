package http

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const PinoBatchJson = `{"logs":[{"level":30,"time":1691719960798,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","msg":"generating sitemap"},{"level":30,"time":1691719961378,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","msg":"got remote data"},{"level":30,"time":1691719961379,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","numPages":91,"msg":"build pages"},{"level":30,"time":1691719965738,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","msg":"generating sitemap"},{"level":30,"time":1691719966256,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","msg":"got remote data"},{"level":30,"time":1691719966256,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","numPages":91,"msg":"build pages"},{"level":30,"time":1691719967152,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","msg":"generating sitemap"},{"level":30,"time":1691719967401,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","msg":"got remote data"},{"level":30,"time":1691719967402,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","numPages":91,"msg":"build pages"},{"level":30,"time":1691719967927,"pid":47069,"hostname":"Vadims-MacBook-Pro.local","msg":"generating sitemap"}]}`

const FlyNDJson = `{"event":{"provider":"app"},"fly":{"app":{"instance":"4d891d77b05118","name":"restless-moon-8905"},"region":"lax"},"host":"f442","log":{"level":"info"},"message":"2023-06-27T01:19:11.788889Z ERROR sink{component_kind=\"sink\" component_id=highlight component_type=http component_name=highlight}:request{request_id=118}: vector::sinks::util::retries: Internal log [Not retriable; dropping the request.] is being rate limited.","timestamp":"2023-06-27T01:19:11.789045558Z"}
{"event":{"provider":"app"},"fly":{"app":{"instance":"4d891d77b05118","name":"restless-moon-8905"},"region":"lax"},"host":"f442","log":{"level":"info"},"message":"2023-06-27T01:19:11.788972Z ERROR sink{component_kind=\"sink\" component_id=highlight component_type=http component_name=highlight}:request{request_id=118}: vector::sinks::util::sink: Response failed. response=Response { status: 400, version: HTTP/1.1, headers: {\"content-length\": \"44\", \"content-type\": \"text/plain; charset=utf-8\", \"date\": \"Tue, 27 Jun 2023 01:19:11 GMT\", \"ngrok-trace-id\": \"6a294abe34d80ed9bd1a4817e091c551\", \"vary\": \"Origin\", \"x-content-type-options\": \"nosniff\"}, body: b\"invalid character '{' after top-level value\\n\" }","timestamp":"2023-06-27T01:19:11.789058388Z"}
{"event":{"provider":"app"},"fly":{"app":{"instance":"4d891d77b05118","name":"restless-moon-8905"},"region":"lax"},"host":"f442","log":{"level":"info"},"message":"2023-06-27T01:19:11.789025Z ERROR sink{component_kind=\"sink\" component_id=highlight component_type=http component_name=highlight}:request{request_id=118}: vector_common::internal_event::service: Internal log [Service call failed. No retries or retries exhausted.] is being rate limited.","timestamp":"2023-06-27T01:19:11.789115699Z"}
{"event":{"provider":"app"},"fly":{"app":{"instance":"4d891d77b05118","name":"restless-moon-8905"},"region":"lax"},"host":"f442","log":{"level":"info"},"message":"2023-06-27T01:19:11.789065Z ERROR sink{component_kind=\"sink\" component_id=highlight component_type=http component_name=highlight}:request{request_id=118}: vector_common::internal_event::component_events_dropped: Internal log [Events dropped] is being rate limited.","timestamp":"2023-06-27T01:19:11.789131359Z"}
{"event":{"provider":"app"},"fly":{"app":{"instance":"4d891d77b05118","name":"restless-moon-8905"},"region":"lax"},"host":"f442","log":{"level":"info"},"message":"2023-06-27T01:19:12.433035Z ERROR sink{component_kind=\"sink\" component_id=highlight component_type=http component_name=highlight}:request{request_id=119}: vector::sinks::util::sink: Response failed. response=Response { status: 400, version: HTTP/1.1, headers: {\"content-length\": \"44\", \"content-type\": \"text/plain; charset=utf-8\", \"date\": \"Tue, 27 Jun 2023 01:19:12 GMT\", \"ngrok-trace-id\": \"40c1f0d311b45ab57e16e9d367628dc3\", \"vary\": \"Origin\", \"x-content-type-options\": \"nosniff\"}, body: b\"invalid character '{' after top-level value\\n\" }","timestamp":"2023-06-27T01:19:12.433213947Z"}
{"event":{"provider":"app"},"fly":{"app":{"instance":"6e82d4e6a37548","name":"fly-builder-autumn-violet-9735"},"region":"lax"},"host":"971e","log":{"level":"info"},"message":"time=\"2023-06-27T01:19:12.541516209Z\" level=debug msg=\"checking docker activity\"","timestamp":"2023-06-27T01:19:12.541906845Z"}
{"event":{"provider":"app"},"fly":{"app":{"instance":"6e82d4e6a37548","name":"fly-builder-autumn-violet-9735"},"region":"lax"},"host":"971e","log":{"level":"info"},"message":"time=\"2023-06-27T01:19:12.541802849Z\" level=debug msg=\"Calling GET /v1.41/containers/json?filters=%7B%22status%22%3A%7B%22running%22%3Atrue%7D%7D&limit=0\"","timestamp":"2023-06-27T01:19:12.542083756Z"}`

const GCPJson = `{"insertId":"o9knqgrvve37m18j","jsonPayload":{"job":"work.email.recurring_queue_new_review_notifications","job-id":"ff0bfe05-d764-4885-adbf-be5c3fd6fa57","job-queue":"qa:kvasir","level":"info","max-retry":5,"msg":"processing task","retry":0,"worker":"asynq"},"labels":{"compute.googleapis.com/resource_name":"gke-staging-spot-pool-4741f477-ts57","k8s-pod/app":"worker","k8s-pod/app_kubernetes_io/managed-by":"shelob","k8s-pod/pod-template-hash":"6cc89b447c","k8s-pod/security_istio_io/tlsMode":"istio","k8s-pod/service_istio_io/canonical-name":"worker","k8s-pod/service_istio_io/canonical-revision":"latest"},"logName":"projects/precisely-staging/logs/stdout","receiveTimestamp":"2024-04-18T11:15:04.985600039Z","resource":{"labels":{"cluster_name":"staging","container_name":"worker","location":"europe-west3-a","namespace_name":"qa","pod_name":"worker-deployment-6cc89b447c-wbksh","project_id":"precisely-staging"},"type":"k8s_container"},"severity":"INFO","timestamp":"2024-04-18T11:15:00Z"}`

const KinesesFirehoseCloudwatch = ""

const KinesisFirehoseFirelensJson = `{
    "source": "stderr",
    "log": "something happened in this execution.",
    "container_id": "b202eacdb71a473e812e81eaf8f4f8c0-1057226457",
    "container_name": "example-json-logger",
    "ecs_cluster": "highlight-production-cluster",
    "ecs_task_arn": "arn:aws:ecs:us-east-2:173971919437:task/highlight-production-cluster/b202eacdb71a473e812e81eaf8f4f8c0",
    "ecs_task_definition": "example-json-logger:3",
  	"highlight.trace_id": "f80fc1e87e7bce2bb992167f47f8ab00"
}`

const KinesisFirehoseFirelensFluentbitJson = `{
    "@timestamp": "2024-10-07T22:18:32+0000",
    "level": "ERROR",
    "message": "something happened in this execution.",
    "source": "stdout",
    "container_id": "20f475b66b2b45c8b4253d9fbf40ff1d-1057226457",
    "container_name": "example-json-logger",
    "ecs_cluster": "highlight-production-cluster",
    "ecs_task_arn": "arn:aws:ecs:us-east-2:173971919437:task/highlight-production-cluster/20f475b66b2b45c8b4253d9fbf40ff1d",
    "ecs_task_definition": "example-json-logger:4",
  	"trace_id": "f80fc1e87e7bce2bb992167f47f8ab00"
}`

const KinesisFirehoseFirelensPinoJson = `{
  "status": "error",
  "level": 50,
  "time": 1728347034087,
  "pid": 43,
  "environment": "staging",
  "region": "us-east-2",
  "name": "example-json-logger",
  "service": "abc123",
  "dd": {
    "trace_id": "7211330732019911779",
    "span_id": "7995688579442497114",
    "service": "main-api",
    "version": "1.2.3",
    "env": "staging"
  },
  "traceId": "67047adc000000007307322150893952",
  "trace_id": "f80fc1e87e7bce2bb992167f47f8ab00",
  "span_id": "9fdca739939f9145",
  "trace_flags": "01",
  "flagKey": "foo-bar-baz",
  "requestContext": {
    "userId": "abc-123",
    "organizationId": 1
  },
  "defaultValue": false,
  "variationValue": false,
  "msg": "something happened in this execution.",
  "container_id": "x-y",
  "container_name": "api-container-name",
  "source": "stdout",
  "ecs_cluster": "fooAPI",
  "ecs_task_arn": "arn:aws:ecs:us-east-2:x:task/fooAPI/y",
  "ecs_task_definition": "fooAPI:659"
}`

const KinesisFirehoseCloudFrontJson = `{"timestamp":"1733943532","DistributionId":"E20MFWZTRJBW2X","date":"2024-12-11","time":"18:58:52","x-edge-location":"IAD55-P7","sc-bytes":"2932","c-ip":"52.71.51.89","cs-method":"POST","cs(Host)":"d3tbhpzcw8lafv.cloudfront.net","cs-uri-stem":"/","sc-status":"200","cs(Referer)":"-","cs(User-Agent)":"python-requests/2.31.0","cs-uri-query":"-","cs(Cookie)":"-","x-edge-result-type":"Miss","x-edge-request-id":"Q37mFQxX4juKEkB4xJ6A8ZYHLtMoO_0_fYyLt0H8vnRJywgocMuFbA==","x-host-header":"pri.highlight.io","cs-protocol":"https","cs-bytes":"1428","time-taken":"0.047","x-forwarded-for":"-","ssl-protocol":"TLSv1.3","ssl-cipher":"TLS_AES_128_GCM_SHA256","x-edge-response-result-type":"Miss","cs-protocol-version":"HTTP/1.1","fle-status":"-","fle-encrypted-fields":"-","c-port":"18366","time-to-first-byte":"0.047","x-edge-detailed-result-type":"Miss","sc-content-type":"application/json","sc-content-len":"-","sc-range-start":"-","sc-range-end":"-","timestamp(ms)":"1733943532406","origin-fbl":"0.042","origin-lbl":"0.042","asn":"14618"}`

const KinesisFirehoseJson = `{"timestamp":1734049709831,"formatVersion":1,"webaclId":"arn:aws:wafv2:us-east-1:173971919437:global/webacl/CreatedByCloudFront-ff696fdc-2ada-4504-a1f8-ee8693866a67/4c7c9db0-8c80-411d-95c0-b54fb42667a8","terminatingRuleId":"Default_Action","terminatingRuleType":"REGULAR","action":"ALLOW","terminatingRuleMatchDetails":[],"httpSourceName":"CF","httpSourceId":"E20MFWZTRJBW2X","ruleGroupList":[{"ruleGroupId":"AWS#AWSManagedRulesAmazonIpReputationList","terminatingRule":null,"nonTerminatingMatchingRules":[],"excludedRules":null,"customerConfig":null},{"ruleGroupId":"AWS#AWSManagedRulesCommonRuleSet","terminatingRule":null,"nonTerminatingMatchingRules":[],"excludedRules":null,"customerConfig":null},{"ruleGroupId":"AWS#AWSManagedRulesKnownBadInputsRuleSet","terminatingRule":null,"nonTerminatingMatchingRules":[],"excludedRules":null,"customerConfig":null},{"ruleGroupId":"AWS#AWSManagedRulesSQLiRuleSet","terminatingRule":null,"nonTerminatingMatchingRules":[],"excludedRules":null,"customerConfig":null}],"rateBasedRuleList":[],"nonTerminatingMatchingRules":[],"requestHeadersInserted":null,"responseCodeSent":null,"httpRequest":{"clientIp":"2403:5808:b0a6:0:2d05:e622:c0b2:5746","country":"AU","headers":[{"name":"host","value":"pri.highlight.io"},{"name":"origin","value":"https://app.highlight.io"},{"name":"sec-fetch-site","value":"same-site"},{"name":"access-control-request-method","value":"POST"},{"name":"access-control-request-headers","value":"content-type,token,traceparent,x-highlight-request"},{"name":"sec-fetch-mode","value":"cors"},{"name":"user-agent","value":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15"},{"name":"referer","value":"https://app.highlight.io/"},{"name":"sec-fetch-dest","value":"empty"},{"name":"content-length","value":"0"},{"name":"accept","value":"*/*"},{"name":"accept-language","value":"en-US,en;q=0.9"},{"name":"priority","value":"u=3, i"},{"name":"accept-encoding","value":"gzip, deflate, br"}],"uri":"/","args":"","httpVersion":"HTTP/2.0","httpMethod":"OPTIONS","requestId":"jCWmLaJY653BbNBZmaW3kE3taUAMakWkk70Q1IH13F1OLO8O9-TAmA=="},"ja3Fingerprint":"773906b0efdefa24a7f2b8eb6985bf37", "level": "warning"}`

type MockResponseWriter struct {
	statusCode int
}

func (m *MockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (m *MockResponseWriter) Write(bytes []byte) (int, error) {
	return 0, nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

var spanRecorder = tracetest.NewSpanRecorder()

func TestMain(m *testing.M) {
	tracer = sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanRecorder),
	).Tracer("test")

	code := m.Run()
	os.Exit(code)
}

func TestHandleRawLog(t *testing.T) {
	r, _ := http.NewRequest("POST", fmt.Sprintf("/v1/logs/raw?%s=1jdkoe52&%s=test", LogDrainProjectQueryParam, LogDrainServiceQueryParam), strings.NewReader("yo there, this is the message"))
	w := &MockResponseWriter{}
	HandleRawLog(w, r)
	assert.Equal(t, 200, w.statusCode)
}

func TestHandleFlyJSONLog(t *testing.T) {
	r, _ := http.NewRequest("POST", "/v1/logs/json", strings.NewReader(FlyNDJson))
	r.Header.Set("Content-Type", "application/x-ndjson")
	r.Header.Set(LogDrainProjectHeader, "1")
	w := &MockResponseWriter{}
	HandleJSONLog(w, r)
	assert.Equal(t, 200, w.statusCode)
}

func TestHandlePinoBatchJson(t *testing.T) {
	r, _ := http.NewRequest("POST", "/v1/logs/json?project=1", strings.NewReader(PinoBatchJson))
	r.Header.Set("Content-Type", "application/json")
	w := &MockResponseWriter{}
	HandleJSONLog(w, r)
	assert.Equal(t, 200, w.statusCode)
}

func TestHandleFlyJSONGZIPLog(t *testing.T) {
	b := bytes.Buffer{}
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(FlyNDJson)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	r, _ := http.NewRequest("POST", "/v1/logs/json", &b)
	r.Header.Set("Content-Type", "application/x-ndjson")
	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set(LogDrainProjectHeader, "1")
	r.Header.Set(LogDrainServiceHeader, "foo")
	w := &MockResponseWriter{}
	HandleJSONLog(w, r)
	assert.Equal(t, 200, w.statusCode)
}

func TestHandleGCPJson(t *testing.T) {
	r, _ := http.NewRequest("POST", "/v1/logs/json?project=1jdkoe52&service=backend-service", strings.NewReader(GCPJson))
	r.Header.Set("Content-Type", "application/json")
	w := &MockResponseWriter{}
	HandleJSONLog(w, r)
	assert.Equal(t, 200, w.statusCode)

	spans := spanRecorder.Ended()
	span := spans[len(spans)-1]
	event := span.Events()[0]
	assert.Equal(t, "highlight.log", span.Name())
	assert.Equal(t, "log", event.Name)

	proj, _ := lo.Find(span.Attributes(), func(item attribute.KeyValue) bool {
		return item.Key == "highlight.project_id"
	})
	assert.Equal(t, "1", proj.Value.AsString())

	serv, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "service.name"
	})
	assert.Equal(t, "backend-service", serv.Value.AsString())

	sev, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "severity"
	})
	assert.Equal(t, "INFO", sev.Value.AsString())

	msg, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "jsonPayload.msg"
	})
	assert.Equal(t, "processing task", msg.Value.AsString())
}

func TestHandleFirehoseCloudwatch(t *testing.T) {
	r, _ := http.NewRequest("POST", "/v1/logs/firehose", bytes.NewReader([]byte(KinesesFirehoseCloudwatch)))
	w := &MockResponseWriter{}
	HandleFirehoseLog(w, r)
	assert.Equal(t, http.StatusBadRequest, w.statusCode)
}

func TestHandleFirehoseFireLens(t *testing.T) {
	for idx, body := range []string{KinesisFirehoseFirelensJson, KinesisFirehoseFirelensFluentbitJson, KinesisFirehoseFirelensPinoJson} {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			b := bytes.Buffer{}
			gz := gzip.NewWriter(&b)
			if _, err := gz.Write([]byte(body)); err != nil {
				t.Fatal(err)
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
			bd := fmt.Sprintf(`{"timestamp":%d,"records":[{"data": "%s"}]}`, time.Now().UTC().UnixMilli(), base64.StdEncoding.EncodeToString(b.Bytes()))
			t.Log("body", bd)

			r, _ := http.NewRequest("POST", "/v1/logs/firehose", strings.NewReader(bd))
			r.Header.Set("X-Amz-Firehose-Common-Attributes", fmt.Sprintf(`{"commonAttributes":{"x-highlight-project":"%d"}}`, 2+idx))
			w := &MockResponseWriter{}
			HandleFirehoseLog(w, r)
			assert.Equal(t, 200, w.statusCode)

			spans := spanRecorder.Ended()
			span := spans[len(spans)-1]
			event := span.Events()[len(span.Events())-1]
			assert.True(t, span.SpanContext().TraceID().IsValid())
			assert.Equal(t, "f80fc1e87e7bce2bb992167f47f8ab00", span.SpanContext().TraceID().String())
			assert.Equal(t, "highlight.log", span.Name())
			assert.Equal(t, "log", event.Name)

			// in the last 30 years to check that time is not unix epoch
			assert.Less(t, time.Since(event.Time), time.Hour*24*365*30)

			proj, _ := lo.Find(span.Attributes(), func(item attribute.KeyValue) bool {
				return item.Key == "highlight.project_id"
			})
			assert.Equal(t, fmt.Sprintf("%d", 2+idx), proj.Value.AsString())

			serv, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
				return item.Key == "service.name"
			})
			assert.Equal(t, "example-json-logger", serv.Value.AsString())

			sev, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
				return item.Key == "log.severity"
			})
			assert.Equal(t, "error", sev.Value.AsString())

			msg, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
				return item.Key == "log.message"
			})
			assert.Equal(t, "something happened in this execution.", msg.Value.AsString())
		})
	}
}

func TestHandleKinesisFirehoseCloudFrontJson(t *testing.T) {
	b := bytes.Buffer{}
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(KinesisFirehoseCloudFrontJson)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	bd := fmt.Sprintf(`{"timestamp":%d,"records":[{"data": "%s"}]}`, time.Now().UTC().UnixMilli(), base64.StdEncoding.EncodeToString(b.Bytes()))
	t.Log("body", bd)

	r, _ := http.NewRequest("POST", "/v1/logs/firehose", strings.NewReader(bd))
	r.Header.Set("X-Amz-Firehose-Common-Attributes", fmt.Sprintf(`{"commonAttributes":{"x-highlight-project":"%d"}}`, 3))
	w := &MockResponseWriter{}
	HandleFirehoseLog(w, r)
	assert.Equal(t, 200, w.statusCode)

	spans := spanRecorder.Ended()
	span := spans[len(spans)-1]
	event := span.Events()[len(span.Events())-1]
	assert.True(t, span.SpanContext().TraceID().IsValid())
	assert.Equal(t, "highlight.log", span.Name())
	assert.Equal(t, "log", event.Name)

	// in the last 30 years to check that time is not unix epoch
	assert.Less(t, time.Since(event.Time), time.Hour*24*365*30)

	proj, _ := lo.Find(span.Attributes(), func(item attribute.KeyValue) bool {
		return item.Key == "highlight.project_id"
	})
	assert.Equal(t, fmt.Sprintf("%d", 3), proj.Value.AsString())

	serv, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "service.name"
	})
	assert.Equal(t, "firehose.cloudfront", serv.Value.AsString())

	sev, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "log.severity"
	})
	assert.Equal(t, "info", sev.Value.AsString())

	msg, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "log.message"
	})
	assert.Equal(t, "[POST 200] d3tbhpzcw8lafv.cloudfront.net/", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "x-host-header"
	})
	assert.Equal(t, "pri.highlight.io", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "cs(Host)"
	})
	assert.Equal(t, "", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "cs.Host"
	})
	assert.Equal(t, "d3tbhpzcw8lafv.cloudfront.net", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "cs(User-Agent)"
	})
	assert.Equal(t, "", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "cs.User-Agent"
	})
	assert.Equal(t, "python-requests/2.31.0", msg.Value.AsString())
}

func TestHandleKinesisFirehoseJson(t *testing.T) {
	b := bytes.Buffer{}
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(KinesisFirehoseJson)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	bd := fmt.Sprintf(`{"timestamp":%d,"records":[{"data": "%s"}]}`, time.Now().UTC().UnixMilli(), base64.StdEncoding.EncodeToString(b.Bytes()))
	t.Log("body", bd)

	r, _ := http.NewRequest("POST", "/v1/logs/firehose", strings.NewReader(bd))
	r.Header.Set("X-Amz-Firehose-Common-Attributes", fmt.Sprintf(`{"commonAttributes":{"x-highlight-project":"%d"}}`, 3))
	w := &MockResponseWriter{}
	HandleFirehoseLog(w, r)
	assert.Equal(t, 200, w.statusCode)

	spans := spanRecorder.Ended()
	span := spans[len(spans)-1]
	event := span.Events()[len(span.Events())-1]
	assert.True(t, span.SpanContext().TraceID().IsValid())
	assert.Equal(t, "highlight.log", span.Name())
	assert.Equal(t, "log", event.Name)

	// in the last 30 years to check that time is not unix epoch
	assert.Less(t, time.Since(event.Time), time.Hour*24*365*30)

	proj, _ := lo.Find(span.Attributes(), func(item attribute.KeyValue) bool {
		return item.Key == "highlight.project_id"
	})
	assert.Equal(t, fmt.Sprintf("%d", 3), proj.Value.AsString())

	serv, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "service.name"
	})
	assert.Equal(t, "firehose.json", serv.Value.AsString())

	sev, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "log.severity"
	})
	assert.Equal(t, "warning", sev.Value.AsString())

	msg, _ := lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "log.message"
	})
	assert.Equal(t, "JSON", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "action"
	})
	assert.Equal(t, "ALLOW", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "httpRequest.clientIp"
	})
	assert.Equal(t, "2403:5808:b0a6:0:2d05:e622:c0b2:5746", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "httpRequest.headers.0.name"
	})
	assert.Equal(t, "host", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "httpRequest.headers.0.value"
	})
	assert.Equal(t, "pri.highlight.io", msg.Value.AsString())

	msg, _ = lo.Find(event.Attributes, func(item attribute.KeyValue) bool {
		return item.Key == "ruleGroupList.0.ruleGroupId"
	})
	assert.Equal(t, "AWS#AWSManagedRulesAmazonIpReputationList", msg.Value.AsString())
}
