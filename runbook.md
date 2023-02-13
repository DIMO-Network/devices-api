# Runbook

## Pod restarting every so often
This could be caused by OOM or an application panic. For the latter, usually checking the previous container's logs may 
help you quickly find the panic causing the problem:
`kc logs --previous -n prod <podname>`

To check why the restarts are happening in the first place:
`kc describe pod -n prod <podname>`
Then look for: Containers -> Last State -> Reason