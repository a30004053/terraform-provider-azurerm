package domains

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type ListBySubscriptionResponse struct {
	HttpResponse *http.Response
	Model        *[]Domain

	nextLink     *string
	nextPageFunc func(ctx context.Context, nextLink string) (ListBySubscriptionResponse, error)
}

type ListBySubscriptionCompleteResult struct {
	Items []Domain
}

func (r ListBySubscriptionResponse) HasMore() bool {
	return r.nextLink != nil
}

func (r ListBySubscriptionResponse) LoadMore(ctx context.Context) (resp ListBySubscriptionResponse, err error) {
	if !r.HasMore() {
		err = fmt.Errorf("no more pages returned")
		return
	}
	return r.nextPageFunc(ctx, *r.nextLink)
}

type ListBySubscriptionOptions struct {
	Filter *string
	Top    *int64
}

func DefaultListBySubscriptionOptions() ListBySubscriptionOptions {
	return ListBySubscriptionOptions{}
}

func (o ListBySubscriptionOptions) toQueryString() map[string]interface{} {
	out := make(map[string]interface{})

	if o.Filter != nil {
		out["$filter"] = *o.Filter
	}

	if o.Top != nil {
		out["$top"] = *o.Top
	}

	return out
}

// ListBySubscription ...
func (c DomainsClient) ListBySubscription(ctx context.Context, id SubscriptionId, options ListBySubscriptionOptions) (resp ListBySubscriptionResponse, err error) {
	req, err := c.preparerForListBySubscription(ctx, id, options)
	if err != nil {
		err = autorest.NewErrorWithError(err, "domains.DomainsClient", "ListBySubscription", nil, "Failure preparing request")
		return
	}

	resp.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "domains.DomainsClient", "ListBySubscription", resp.HttpResponse, "Failure sending request")
		return
	}

	resp, err = c.responderForListBySubscription(resp.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "domains.DomainsClient", "ListBySubscription", resp.HttpResponse, "Failure responding to request")
		return
	}
	return
}

// ListBySubscriptionComplete retrieves all of the results into a single object
func (c DomainsClient) ListBySubscriptionComplete(ctx context.Context, id SubscriptionId, options ListBySubscriptionOptions) (ListBySubscriptionCompleteResult, error) {
	return c.ListBySubscriptionCompleteMatchingPredicate(ctx, id, options, DomainPredicate{})
}

// ListBySubscriptionCompleteMatchingPredicate retrieves all of the results and then applied the predicate
func (c DomainsClient) ListBySubscriptionCompleteMatchingPredicate(ctx context.Context, id SubscriptionId, options ListBySubscriptionOptions, predicate DomainPredicate) (resp ListBySubscriptionCompleteResult, err error) {
	items := make([]Domain, 0)

	page, err := c.ListBySubscription(ctx, id, options)
	if err != nil {
		err = fmt.Errorf("loading the initial page: %+v", err)
		return
	}
	if page.Model != nil {
		for _, v := range *page.Model {
			if predicate.Matches(v) {
				items = append(items, v)
			}
		}
	}

	for page.HasMore() {
		page, err = page.LoadMore(ctx)
		if err != nil {
			err = fmt.Errorf("loading the next page: %+v", err)
			return
		}

		if page.Model != nil {
			for _, v := range *page.Model {
				if predicate.Matches(v) {
					items = append(items, v)
				}
			}
		}
	}

	out := ListBySubscriptionCompleteResult{
		Items: items,
	}
	return out, nil
}

// preparerForListBySubscription prepares the ListBySubscription request.
func (c DomainsClient) preparerForListBySubscription(ctx context.Context, id SubscriptionId, options ListBySubscriptionOptions) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	for k, v := range options.toQueryString() {
		queryParameters[k] = autorest.Encode("query", v)
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/providers/Microsoft.EventGrid/domains", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// preparerForListBySubscriptionWithNextLink prepares the ListBySubscription request with the given nextLink token.
func (c DomainsClient) preparerForListBySubscriptionWithNextLink(ctx context.Context, nextLink string) (*http.Request, error) {
	uri, err := url.Parse(nextLink)
	if err != nil {
		return nil, fmt.Errorf("parsing nextLink %q: %+v", nextLink, err)
	}
	queryParameters := map[string]interface{}{}
	for k, v := range uri.Query() {
		if len(v) == 0 {
			continue
		}
		val := v[0]
		val = autorest.Encode("query", val)
		queryParameters[k] = val
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(uri.Path),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForListBySubscription handles the response to the ListBySubscription request. The method always
// closes the http.Response Body.
func (c DomainsClient) responderForListBySubscription(resp *http.Response) (result ListBySubscriptionResponse, err error) {
	type page struct {
		Values   []Domain `json:"value"`
		NextLink *string  `json:"nextLink"`
	}
	var respObj page
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&respObj),
		autorest.ByClosing())
	result.HttpResponse = resp
	result.Model = &respObj.Values
	result.nextLink = respObj.NextLink
	if respObj.NextLink != nil {
		result.nextPageFunc = func(ctx context.Context, nextLink string) (result ListBySubscriptionResponse, err error) {
			req, err := c.preparerForListBySubscriptionWithNextLink(ctx, nextLink)
			if err != nil {
				err = autorest.NewErrorWithError(err, "domains.DomainsClient", "ListBySubscription", nil, "Failure preparing request")
				return
			}

			result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
			if err != nil {
				err = autorest.NewErrorWithError(err, "domains.DomainsClient", "ListBySubscription", result.HttpResponse, "Failure sending request")
				return
			}

			result, err = c.responderForListBySubscription(result.HttpResponse)
			if err != nil {
				err = autorest.NewErrorWithError(err, "domains.DomainsClient", "ListBySubscription", result.HttpResponse, "Failure responding to request")
				return
			}

			return
		}
	}
	return
}