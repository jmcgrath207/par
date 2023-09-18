

* Webhooks needs to know about meta information about all records, currently a single deployment controller is started to bound to that information. This is important so we can filter and apply the correct updates.
  * This information includes:
    * Namepaces
    * Labels
    * ID of the DNS client
    * DNS address

* With this Cached information will look up a hash based on Namespace + Labels
  * This will return
    * ID of the client
    * DNS Address