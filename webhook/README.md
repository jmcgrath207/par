

Added code patch's need to trouble crashing issue.


Flow:
* Move setManagerIPaddress to Resources and start first
* Trigger backfill of deployments (make async job that runs every 2 minutes after start)
  * If deployment has create DNS IP address, Ignore
  * If Deployment has None or Wrong DNS IP address.
    * Trigger update so it can by processed by webhook.





// Find existing deployments, add a label patch of par: procssing, then remove label in the webhook.
// 	err := store.ClientK8s.Patch(context.TODO(), deploymentClone, client.MergeFrom(&deployment))

