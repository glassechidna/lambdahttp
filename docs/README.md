# lambdahttp

Do you yearn for the days of listening on port 80(80) for web traffic? Does this
new serverless world scare you and make you wonder if you'll have to refactor
your existing code? Or maybe you have a crusty old binary lying around, have
lost the source code and are desperate to make it work on AWS Lambda?

If so, `lambdahttp` is for **you**.

## How do I use it?

* Grab the latest zip from the [Releases][releases]
  tab on GitHub.
* Add your app to the zip file.
* Upload the zip to S3.
* Create a Lambda function from that zip file. Configure it to use a "custom runtime"
  and specify a shell command as the "handler", e.g. `./myprogram listen --port 8080`.
* Do the API Gateway / Application Load Balancer dance.
* Swim in the savings of serverless.

If you app _doesn't_ listen on port 8080, you can specify a different one using
the `PORT` environment variable. 

Additionally, `lambdahttp` has no real way of knowing when your app is ready to 
start serving traffic. For this reason, it will continuously make requests to 
`/ping` until that endpoint returns a `200 OK` - this is how it knows you are 
good to go. If that path doesn't work for you, specify a different one in the
`HEALTHCHECK_PATH` environment variable.

## How does it work?

Let me explain through the only kind of diagram I know: a sequence diagram.

![seq-diag](/docs/seq-diag.png)

`lambdahttp` communicates with the Lambda service using the [Lambda runtime interface][runtime]
and converts these requests into regular HTTP over TCP. It converts the responses back to the
format expected by Lambda and voilà. Nothing new you need to learn.

## What bonus functionality exists?

Glad you asked. Since Lambda doesn't have the same level of support for Secrets
Manager and Parameter Store that ECS has, `lambdahttp` fills in the gaps. If
you set an environment variable named `EXAMPLE` to value `{aws-ssm}/path/to/param`,
then `lambdahttp` will retrieve the value at `/path/to/param` and populate `EXAMPLE`
with that value. 

Likewise, `EXAMPLE={aws-sm}secretNameOrArn` will cause `lambdahttp` to replace 
the value with the JSON secret in `secretNameOrArn`. `lambdahttp` also 
recognises `{aws-sm}secretNameOrArn::password` syntax, in which case it will
retrieve the secret and populate `EXAMPLE` with only the value in the `password` 
key.

[releases]: https://github.com/glassechidna/lambdahttp/releases
[runtime]: https://docs.aws.amazon.com/lambda/latest/dg/runtimes-api.html
